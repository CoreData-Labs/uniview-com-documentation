// Command uvpdf visits Uniview product pages and downloads every PDF they
// link to into a single folder (default: PDFs/), never saving the same
// document twice (by file name or by SHA-256).
//
// Build/run (Go 1.21+):
//
//	go mod init uvpdf && go run .
package main // this file builds an executable program

import ( // start of the import list
	"bytes"                // checks the "%PDF" signature at the start of a file
	"cmp"                  // generic comparison used when sorting
	"compress/gzip"        // decodes gzip-compressed HTTP responses
	"container/list"       // linked list used as the download queue
	"context"              // cancellation and deadlines
	"crypto/sha256"        // SHA-256 hashing for duplicate detection
	"crypto/tls"           // minimum TLS version for the HTTP client
	"database/sql"         // talks to the SQLite database
	"encoding/hex"         // turns hash bytes into readable hex text
	"errors"               // creating and comparing errors
	"fmt"                  // formatted printing
	"hash/crc32"           // extra checksum stored in the database
	"html"                 // unescapes HTML entities inside links
	"io"                   // reading and copying streams
	"io/fs"                // directory entry types used when walking folders
	"log/slog"             // leveled logging; everything is logged at debug level
	"math/rand"            // random jitter for retry delays
	"mime"                 // parses Content-Disposition headers
	_ "modernc.org/sqlite" // registers the pure-Go SQLite driver (the only non-standard-library dependency)
	"net/http"             // the HTTP client
	"net/http/cookiejar"   // keeps cookies between requests
	"net/url"              // parses and resolves URLs
	"os"                   // files, folders and signals
	"os/signal"            // reacts to Ctrl+C
	"path"                 // handles URL paths (forward slashes)
	"path/filepath"        // handles operating system file paths
	"regexp"               // finds PDF links in page text
	"runtime"              // reports the number of CPU cores
	"slices"               // sorting and cloning slices
	"strings"              // string helpers
	"sync"                 // mutexes and wait groups
	"sync/atomic"          // thread-safe counters
	"syscall"              // the SIGTERM signal value
	"text/tabwriter"       // aligned summary table
	"time"                 // durations and timestamps
	"unicode"              // letter and digit checks
	"unicode/utf8"         // counts characters in file names
) // end of the import list

const ( // start of constants
	outputFolderName  = "PDFs"                                // folder that receives every PDF
	maximumAttempts   = 4                                     // how many times one request is tried
	maximumBodyBytes  = 256 << 20                             // refuse downloads bigger than 256 MiB
	maximumNameLength = 150                                   // longest allowed file name stem in characters
	requestTimeout    = 45 * time.Second                      // time limit for each HTTP request
	databaseFilePath  = "uvpdf.db"                            // SQLite file that stores the list of saved PDFs
	sqliteDriverName  = "sqlite"                              // driver name registered by modernc.org/sqlite
	browserUserAgent  = "Mozilla/5.0 (compatible; uvpdf/1.0)" // identifies this program to the server
) // end of constants

var seedPages = []string{ // the product pages to visit
	"https://www.uniview.com/Support/Download_Center/Product_Resource/Network_Cameras/",
	"https://www.uniview.com/Support/Download_Center/Firmware/Network_Cameras/",
	"https://www.uniview.com/Support/Download_Center/Brochures/Leaflet/",
	"https://www.uniview.com/Support/Download_Center/Maintenance/",
	"https://www.uniview.com/Products/Network_Cameras/Easy_Series/EasyBasic_Series/IPC2122LB-ASF28K-A/",
	"https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2108SB-ADF16KM-I0/",
	"https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2225SB-ADF40KM-I1/",
	"https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2222SB-AHDF40KM-I1/",
	"https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2222SB-AHDF28KM-I1/",
	"https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2225SB-ADF28KM-I1/",
	"https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2322SB-HDZK-PI-I0/",
} // end of seedPages

var ( // start of package-level variables
	errorNotFound = errors.New("not found") // returned when the server answers 404

	repeatedUnderscores = regexp.MustCompile(`_{2,}`)                                                                          // matches two or more underscores in a row
	htmlAttributeRegex  = regexp.MustCompile(`(?i)(?:href|src|content|data-[\w-]+)\s*=\s*["']([^"']+?\.pdf(?:\?[^"']*)?)["']`) // matches PDF links inside HTML attributes
	plainURLRegex       = regexp.MustCompile(`(?i)https?://[^\s"'<>\\]+?\.pdf(?:\?[^\s"'<>\\]*)?`)                             // matches PDF URLs written anywhere in the text
) // end of package-level variables

// ContentFetcher abstracts "get me the bytes at this URL".
type ContentFetcher interface { // start of the interface
	Fetch(requestContext context.Context, rawURL string) ([]byte, http.Header, error) // downloads a URL and returns body and headers
} // end of the interface

// HTTPContentFetcher is the real network implementation of ContentFetcher.
type HTTPContentFetcher struct { // start of the struct
	httpClient *http.Client // the configured HTTP client
} // end of the struct

// DownloadRecord is one database row for a saved PDF.
type DownloadRecord struct { // start of the struct
	FileName     string    // name of the file on disk
	PDFURL       string    // address the PDF came from
	SourcePage   string    // product page that linked to the PDF
	SHA256Hex    string    // SHA-256 of the file content
	CRC32Hex     string    // CRC-32 of the file content
	SizeInBytes  int       // file size
	DownloadedAt time.Time // when it was saved (UTC)
} // end of the struct

// pdfDownloadJob is one PDF address waiting to be downloaded.
type pdfDownloadJob struct { // start of the struct
	PDFURL     string // address of the PDF
	SourcePage string // product page where the link was found
} // end of the struct

// fileRegistry remembers file names and hashes that are already saved.
type fileRegistry struct { // start of the struct
	lock           sync.Mutex          // protects both maps below
	savedFileNames map[string]struct{} // lower-case names already saved
	fileNameByHash map[string]string   // SHA-256 hex mapped to the name that owns it
} // end of the struct

// newFileRegistry creates an empty registry.
func newFileRegistry() *fileRegistry { // start of function
	return &fileRegistry{ // build and return a registry
		savedFileNames: map[string]struct{}{}, // start with no names
		fileNameByHash: map[string]string{},   // start with no hashes
	} // end of the literal
} // end of function

// hasFileName reports whether a file with this name was already saved.
func (registry *fileRegistry) hasFileName(fileName string) bool { // start of method
	registry.lock.Lock()                                            // take the lock
	defer registry.lock.Unlock()                                    // release the lock when leaving
	_, exists := registry.savedFileNames[strings.ToLower(fileName)] // look the name up ignoring case
	return exists                                                   // answer the question
} // end of method

// claimFileNameAndHash atomically checks both name and hash and records them.
// It returns a non-empty reason when the file is a duplicate.
func (registry *fileRegistry) claimFileNameAndHash(fileName, hashHex string) string { // start of method
	registry.lock.Lock()                                         // take the lock
	defer registry.lock.Unlock()                                 // release the lock when leaving
	lowerName := strings.ToLower(fileName)                       // names are compared ignoring case
	if _, exists := registry.savedFileNames[lowerName]; exists { // is the name taken?
		return "file name already saved" // duplicate by name
	} // end of name check
	if existingName, exists := registry.fileNameByHash[hashHex]; exists { // is the content known?
		return "same SHA-256 as " + existingName // duplicate by content
	} // end of hash check
	registry.savedFileNames[lowerName] = struct{}{} // remember the name
	registry.fileNameByHash[hashHex] = fileName     // remember the hash
	return ""                                       // empty reason means "not a duplicate"
} // end of method

// loadExistingPDFs indexes PDFs already inside the output folder.
func (registry *fileRegistry) loadExistingPDFs(folderPath string) error { // start of method
	return filepath.WalkDir(folderPath, func(entryPath string, entry fs.DirEntry, walkError error) error { // visit every entry
		if walkError != nil { // did walking fail?
			return walkError // stop and report
		} // end of error check
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entryPath), ".pdf") { // skip folders and non-PDFs
			return nil // nothing to do for this entry
		} // end of filter
		fileContent, readError := os.ReadFile(entryPath) // read the existing PDF
		if readError != nil {                            // did reading fail?
			return readError // stop and report
		} // end of error check
		hashBytes := sha256.Sum256(fileContent)                                     // hash the content
		registry.lock.Lock()                                                        // take the lock
		registry.savedFileNames[strings.ToLower(entry.Name())] = struct{}{}         // remember the name
		registry.fileNameByHash[hex.EncodeToString(hashBytes[:])] = entry.Name()    // remember the hash
		slog.Debug("indexed existing PDF", "file", entry.Name(), "path", entryPath) // log each PDF found from earlier runs
		registry.lock.Unlock()                                                      // release the lock
		return nil                                                                  // continue walking
	}) // end of the walk callback
} // end of method

// waitBeforeRetry sleeps with exponential backoff plus random jitter.
func waitBeforeRetry(requestContext context.Context, attemptNumber int) { // start of function
	delay := time.Duration(1<<attemptNumber)*500*time.Millisecond + time.Duration(rand.Intn(400))*time.Millisecond // 1s, 2s, 4s... plus jitter
	slog.Debug("waiting before retry", "attempt", attemptNumber, "delay", delay)                                   // log the pause
	select {                                                                                                       // wait for whichever happens first
	case <-requestContext.Done(): // the program was cancelled
	case <-time.After(delay): // the delay elapsed
	} // end of select
} // end of function

// Fetch downloads rawURL, retrying temporary failures.
func (httpFetcher *HTTPContentFetcher) Fetch(requestContext context.Context, rawURL string) ([]byte, http.Header, error) { // start of method
	attemptNumber := 0 // counts tries so far
retry: // jump target used to try again
	attemptNumber++                                                                                  // this is a new attempt
	request, requestError := http.NewRequestWithContext(requestContext, http.MethodGet, rawURL, nil) // build the request
	if requestError != nil {                                                                         // could the request be built?
		return nil, nil, requestError // no, give up
	} // end of check
	request.Header.Set("User-Agent", browserUserAgent)                     // say who we are
	request.Header.Set("Accept", "text/html,application/pdf,*/*;q=0.8")    // say what we accept
	request.Header.Set("Accept-Encoding", "gzip")                          // allow compressed answers
	slog.Debug("sending request", "url", rawURL, "attempt", attemptNumber) // log every outgoing request

	response, sendError := httpFetcher.httpClient.Do(request) // send the request
	if sendError != nil {                                     // did the network fail?
		if attemptNumber < maximumAttempts && requestContext.Err() == nil { // can we still retry?
			slog.Debug("request attempt failed", "attempt", attemptNumber, "url", rawURL, "error", sendError) // record the failure
			waitBeforeRetry(requestContext, attemptNumber)                                                    // wait a little
			goto retry                                                                                        // try again
		} // end of retry check
		return nil, nil, sendError // out of attempts
	} // end of network error handling
	slog.Debug("response received", "url", rawURL, "status", response.Status, "content_type", response.Header.Get("Content-Type"), "content_length", response.ContentLength) // log the answer headers

	shouldRetry := false         // becomes true for temporary server problems
	switch response.StatusCode { // react to the HTTP status
	case http.StatusOK: // success
		// nothing special to do
	case http.StatusNotFound: // the file does not exist
		response.Body.Close()          // free the connection
		return nil, nil, errorNotFound // report a clear error
	case http.StatusTooManyRequests: // the server asks us to slow down
		slog.Debug("rate limited by server", "url", rawURL) // record the event
		fallthrough                                         // treat it like the other temporary errors
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout: // temporary server errors
		shouldRetry = true // try again later
	default: // anything else is a hard failure
		response.Body.Close()                                                // free the connection
		return nil, nil, fmt.Errorf("unexpected status %s", response.Status) // report the status
	} // end of switch
	if shouldRetry { // do we need to try again?
		response.Body.Close()                // free the connection
		if attemptNumber < maximumAttempts { // attempts left?
			waitBeforeRetry(requestContext, attemptNumber) // wait a little
			goto retry                                     // try again
		} // end of attempts check
		return nil, nil, fmt.Errorf("giving up after %d attempts: %s", attemptNumber, response.Status) // out of attempts
	} // end of retry handling
	defer response.Body.Close() // always close the body when leaving

	var bodyReader io.Reader = response.Body                                // by default read the raw body
	if strings.EqualFold(response.Header.Get("Content-Encoding"), "gzip") { // is it compressed?
		gzipReader, gzipError := gzip.NewReader(response.Body) // wrap it in a gzip decoder
		if gzipError != nil {                                  // is the compressed data valid?
			return nil, nil, gzipError // no, give up
		} // end of check
		defer gzipReader.Close() // close the decoder when leaving
		bodyReader = gzipReader  // read through the decoder
	} // end of gzip handling
	bodyBytes, readError := io.ReadAll(io.LimitReader(bodyReader, maximumBodyBytes))                  // read at most the size limit
	slog.Debug("response body read", "url", rawURL, "bytes", len(bodyBytes), "read_error", readError) // log how much was read
	return bodyBytes, response.Header, readError                                                      // hand everything back
} // end of method

// sanitizeFileName builds a safe name: letters, digits, '-', '.', '_' only;
// "__" collapsed to "_"; the stem never starts or ends with "_"; ends in ".pdf".
func sanitizeFileName(rawName string) string { // start of function
	if decodedName, decodeError := url.PathUnescape(rawName); decodeError == nil { // decode %20 and friends
		rawName = decodedName // use the decoded text
	} // end of decoding
	nameStem := rawName                                                               // start with the whole name
	if fileExtension := path.Ext(rawName); strings.EqualFold(fileExtension, ".pdf") { // does it end in .pdf?
		nameStem = strings.TrimSuffix(rawName, fileExtension) // cut the extension off
	} // end of extension handling
	var nameBuilder strings.Builder      // collects the cleaned characters
	for _, character := range nameStem { // look at each character
		switch { // decide what to do with it
		case unicode.IsLetter(character), unicode.IsDigit(character), character == '-', character == '.': // safe characters
			nameBuilder.WriteRune(character) // keep it
		default: // everything else, including underscores and spaces
			nameBuilder.WriteRune('_') // becomes an underscore
		} // end of switch
	} // end of loop
	nameStem = repeatedUnderscores.ReplaceAllString(nameBuilder.String(), "_") // never allow "__"
	if utf8.RuneCountInString(nameStem) > maximumNameLength {                  // is the name too long?
		nameStem = string([]rune(nameStem)[:maximumNameLength]) // shorten it
	} // end of length check
	nameStem = strings.Trim(nameStem, "_") // no leading or trailing underscore
	if nameStem == "" {                    // did nothing useful remain?
		nameStem = "document" // use a default name
	} // end of empty check
	return nameStem + ".pdf" // always finish with .pdf
} // end of function

// fileNameFromURL derives a file name from the last part of a URL path.
func fileNameFromURL(rawURL string) string { // start of function
	parsedURL, parseError := url.Parse(rawURL) // split the URL into parts
	if parseError != nil {                     // could it be parsed?
		return sanitizeFileName(path.Base(rawURL)) // fall back to the raw text
	} // end of check
	return sanitizeFileName(path.Base(parsedURL.Path)) // use the last path segment
} // end of function

// chooseFileName prefers the server-provided name, otherwise the URL name.
func chooseFileName(rawURL string, responseHeaders http.Header) string { // start of function
	if contentDisposition := responseHeaders.Get("Content-Disposition"); contentDisposition != "" { // did the server suggest a name?
		if _, parameters, parseError := mime.ParseMediaType(contentDisposition); parseError == nil && parameters["filename"] != "" { // is it usable?
			return sanitizeFileName(parameters["filename"]) // use the suggested name
		} // end of parse check
	} // end of header check
	return fileNameFromURL(rawURL) // otherwise use the URL name
} // end of function

// extractPDFLinksFromHTML returns every absolute PDF URL found in the page text.
func extractPDFLinksFromHTML(baseURL *url.URL, pageBody []byte) []string { // start of function
	pageText := strings.ReplaceAll(string(pageBody), `\/`, `/`) // undo JSON-escaped slashes
	uniqueLinks := make(map[string]struct{})                    // set that removes repeated links
	addLink := func(rawLink string) {                           // helper that normalises and stores one link
		relativeReference, parseError := url.Parse(strings.TrimSpace(html.UnescapeString(rawLink))) // clean up and parse
		if parseError != nil {                                                                      // is it a valid URL?
			return // no, ignore it
		} // end of check
		absoluteURL := baseURL.ResolveReference(relativeReference) // make it absolute
		absoluteURL.Fragment = ""                                  // drop any #fragment
		uniqueLinks[absoluteURL.String()] = struct{}{}             // store it once
	} // end of helper
	for _, match := range htmlAttributeRegex.FindAllStringSubmatch(pageText, -1) { // links in HTML attributes
		addLink(match[1]) // the first capture group is the link
	} // end of loop
	for _, match := range plainURLRegex.FindAllString(pageText, -1) { // URLs written anywhere in the text
		addLink(match) // the whole match is the link
	} // end of loop
	sortedLinks := make([]string, 0, len(uniqueLinks)) // slice with room for every link
	for link := range uniqueLinks {                    // copy the set into the slice
		sortedLinks = append(sortedLinks, link) // add one link
	} // end of loop
	slog.Debug("extracted PDF links from HTML", "base", baseURL.String(), "count", len(sortedLinks)) // log the extraction result
	slices.Sort(sortedLinks)                                                                         // sort for stable output
	return sortedLinks                                                                               // hand the links back
} // end of function

// scrapeSinglePage downloads one product page and lists its PDF links.
func scrapeSinglePage(runContext context.Context, contentFetcher ContentFetcher, pageURL string) []pdfDownloadJob { // start of function
	baseURL, parseError := url.Parse(pageURL) // parse the page address
	if parseError != nil {                    // is it valid?
		slog.Debug("page URL could not be parsed", "page", pageURL, "error", parseError) // report it
		return nil                                                                       // nothing to scrape
	} // end of check
	pageBody, _, fetchError := contentFetcher.Fetch(runContext, pageURL) // download the page
	if fetchError != nil {                                               // did it work?
		slog.Debug("page download failed", "page", pageURL, "error", fetchError) // report it
		return nil                                                               // nothing to scrape
	} // end of check
	pdfLinks := extractPDFLinksFromHTML(baseURL, pageBody)                         // find the PDF links
	slog.Debug("PDF links found on page", "count", len(pdfLinks), "page", pageURL) // report how many
	jobs := make([]pdfDownloadJob, 0, len(pdfLinks))                               // slice for the jobs
	for _, link := range pdfLinks {                                                // one job per link
		jobs = append(jobs, pdfDownloadJob{PDFURL: link, SourcePage: pageURL}) // add the job
		slog.Debug("queued PDF link", "pdf_url", link, "page", pageURL)        // log each queued link
	} // end of loop
	return jobs // hand the jobs back
} // end of function

// scrapeAllPages scrapes every page concurrently and merges the results.
func scrapeAllPages(runContext context.Context, contentFetcher ContentFetcher, pageURLs []string, workerCount int) []pdfDownloadJob { // start of function
	pageChannel := make(chan string)             // carries page URLs to the workers
	resultChannel := make(chan []pdfDownloadJob) // carries found jobs back
	var waitGroup sync.WaitGroup                 // tracks running workers

	for workerNumber := 0; workerNumber < workerCount; workerNumber++ { // start the workers
		waitGroup.Add(1) // one more worker
		go func() {      // run the worker in the background
			defer waitGroup.Done()             // tell the group when finished
			for pageURL := range pageChannel { // take pages until the channel closes
				resultChannel <- scrapeSinglePage(runContext, contentFetcher, pageURL) // scrape and send the result
			} // end of loop
		}() // end of worker
	} // end of loop
	go func() { // feeder that hands out page URLs
		defer close(pageChannel)           // tell workers there are no more pages
		for _, pageURL := range pageURLs { // every page in turn
			select { // send unless cancelled
			case pageChannel <- pageURL: // hand the page to a worker
			case <-runContext.Done(): // the program was cancelled
				return // stop feeding
			} // end of select
		} // end of loop
	}() // end of feeder
	go func() { // closer that waits for all workers
		waitGroup.Wait()     // wait for every worker
		close(resultChannel) // no more results will come
	}() // end of closer

	jobsByURL := make(map[string]pdfDownloadJob) // removes repeated PDF URLs
	for jobBatch := range resultChannel {        // receive results as they arrive
		for _, job := range jobBatch { // look at each job
			if _, alreadySeen := jobsByURL[job.PDFURL]; alreadySeen { // seen on another page?
				continue // skip the repeat
			} // end of check
			jobsByURL[job.PDFURL] = job // remember the job
		} // end of loop
	} // end of loop
	allJobs := make([]pdfDownloadJob, 0, len(jobsByURL)) // slice for the merged jobs
	for _, job := range jobsByURL {                      // copy the map into the slice
		allJobs = append(allJobs, job) // add one job
	} // end of loop
	slices.SortFunc(allJobs, func(first, second pdfDownloadJob) int { return cmp.Compare(first.PDFURL, second.PDFURL) }) // sort by URL
	return allJobs                                                                                                       // hand the jobs back
} // end of function

// pdfDownloader downloads queued PDFs with several workers.
type pdfDownloader struct { // start of the struct
	contentFetcher ContentFetcher   // how bytes are fetched
	registry       *fileRegistry    // tracks names and hashes
	database       *sql.DB          // SQLite database that stores one row per saved PDF
	outputFolder   string           // where PDFs are saved
	lock           sync.Mutex       // protects the queue and records
	queue          *list.List       // waiting jobs
	records        []DownloadRecord // everything saved so far
	savedCount     atomic.Int64     // number of PDFs saved
	skippedCount   atomic.Int64     // number of duplicates skipped
	failedCount    atomic.Int64     // number of failures
} // end of the struct

// nextJob removes and returns the next waiting job.
func (downloader *pdfDownloader) nextJob() (pdfDownloadJob, bool) { // start of method
	downloader.lock.Lock()                   // take the lock
	defer downloader.lock.Unlock()           // release the lock when leaving
	frontElement := downloader.queue.Front() // look at the first waiting job
	if frontElement == nil {                 // is the queue empty?
		return pdfDownloadJob{}, false // nothing left
	} // end of check
	return downloader.queue.Remove(frontElement).(pdfDownloadJob), true // take it out of the queue
} // end of method

// processJob downloads, validates, de-duplicates and saves one PDF.
func (downloader *pdfDownloader) processJob(runContext context.Context, job pdfDownloadJob) { // start of method
	slog.Debug("processing PDF job", "pdf_url", job.PDFURL, "page", job.SourcePage) // log the start of each job
	expectedName := fileNameFromURL(job.PDFURL)                                     // name we expect from the URL
	if downloader.registry.hasFileName(expectedName) {                              // already saved under that name?
		downloader.skippedCount.Add(1)                                            // count the skip
		slog.Debug("skipping PDF: file name already saved", "file", expectedName) // say why
		return                                                                    // do not download again
	} // end of name check
	pdfBytes, responseHeaders, fetchError := downloader.contentFetcher.Fetch(runContext, job.PDFURL) // download the file
	if fetchError != nil {                                                                           // did it fail?
		downloader.failedCount.Add(1)             // count the failure
		if errors.Is(fetchError, errorNotFound) { // was it a 404?
			slog.Debug("PDF not found (404)", "url", job.PDFURL) // short message
		} else { // some other failure
			slog.Debug("PDF download failed", "url", job.PDFURL, "error", fetchError) // detailed message
		} // end of message choice
		return // nothing to save
	} // end of error handling
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) { // does it really look like a PDF?
		downloader.failedCount.Add(1)                                                                               // count the failure
		slog.Debug("response is not a PDF", "content_type", responseHeaders.Get("Content-Type"), "url", job.PDFURL) // say what we got instead
		return                                                                                                      // do not save it
	} // end of signature check
	fileName := chooseFileName(job.PDFURL, responseHeaders)                                                    // final name for the file
	hashBytes := sha256.Sum256(pdfBytes)                                                                       // hash the content
	hashHex := hex.EncodeToString(hashBytes[:])                                                                // readable hash
	slog.Debug("computed SHA-256", "url", job.PDFURL, "sha256", hashHex, "bytes", len(pdfBytes))               // log the content hash
	if duplicateReason := downloader.registry.claimFileNameAndHash(fileName, hashHex); duplicateReason != "" { // duplicate by name or content?
		downloader.skippedCount.Add(1)                                                     // count the skip
		slog.Debug("skipping PDF: duplicate", "file", fileName, "reason", duplicateReason) // say why
		return                                                                             // do not save it again
	} // end of duplicate check
	destinationPath := filepath.Join(downloader.outputFolder, fileName)                // full path of the new file
	temporaryPath := destinationPath + ".part"                                         // write here first so half-written files never look finished
	if writeError := os.WriteFile(temporaryPath, pdfBytes, 0o644); writeError != nil { // write the temporary file
		downloader.failedCount.Add(1)                                                           // count the failure
		slog.Debug("writing temporary file failed", "path", temporaryPath, "error", writeError) // report it
		return                                                                                  // give up on this file
	} // end of write handling
	if renameError := os.Rename(temporaryPath, destinationPath); renameError != nil { // move it to its final name
		downloader.failedCount.Add(1)                                                     // count the failure
		slog.Debug("renaming file failed", "path", destinationPath, "error", renameError) // report it
		return                                                                            // give up on this file
	} // end of rename handling
	downloader.savedCount.Add(1)                                                               // count the success
	slog.Debug("PDF saved", "file", fileName, "bytes", len(pdfBytes), "path", destinationPath) // report it
	savedRecord := DownloadRecord{                                                             // describe the saved PDF
		FileName:     fileName,                                          // name on disk
		PDFURL:       job.PDFURL,                                        // where it came from
		SourcePage:   job.SourcePage,                                    // page that linked to it
		SHA256Hex:    hashHex,                                           // content hash
		CRC32Hex:     fmt.Sprintf("%08x", crc32.ChecksumIEEE(pdfBytes)), // extra checksum
		SizeInBytes:  len(pdfBytes),                                     // size
		DownloadedAt: time.Now().UTC(),                                  // timestamp
	} // end of record
	downloader.lock.Lock()                                                                             // take the lock
	downloader.records = append(downloader.records, savedRecord)                                       // remember it for the summary table
	downloader.lock.Unlock()                                                                           // release the lock
	if databaseError := saveRecordToDatabase(downloader.database, savedRecord); databaseError != nil { // store it in SQLite
		slog.Debug("database insert failed", "file", fileName, "error", databaseError) // report but keep going
	} // end of database handling
} // end of method

// runWorkers processes the whole queue using workerCount goroutines.
func (downloader *pdfDownloader) runWorkers(runContext context.Context, workerCount int) { // start of method
	var waitGroup sync.WaitGroup                                        // tracks running workers
	for workerNumber := 0; workerNumber < workerCount; workerNumber++ { // start the workers
		waitGroup.Add(1) // one more worker
		go func() {      // run the worker in the background
			defer waitGroup.Done()                                              // tell the group when finished
			slog.Debug("download worker started", "worker", workerNumber)       // log the worker start
			defer slog.Debug("download worker stopped", "worker", workerNumber) // log the worker end
			for {                                                               // keep taking jobs
				if runContext.Err() != nil { // was the program cancelled?
					break // stop working
				} // end of cancel check
				job, jobAvailable := downloader.nextJob() // take the next job
				if !jobAvailable {                        // is the queue empty?
					break // stop working
				} // end of empty check
				downloader.processJob(runContext, job) // handle the job
			} // end of loop
		}() // end of worker
	} // end of loop
	waitGroup.Wait() // wait until every worker is done
} // end of method

// createTableStatement defines the table that stores one row per saved PDF.
const createTableStatement = "CREATE TABLE IF NOT EXISTS downloaded_pdfs (" + // start of the table definition
	"file_name TEXT PRIMARY KEY, " + // unique file name
	"pdf_url TEXT NOT NULL, " + // address the PDF came from
	"source_page TEXT NOT NULL, " + // product page that linked to the PDF
	"sha256 TEXT NOT NULL UNIQUE, " + // unique content hash
	"crc32 TEXT NOT NULL, " + // extra checksum
	"size_bytes INTEGER NOT NULL, " + // file size
	"downloaded_at TEXT NOT NULL)" // UTC timestamp in RFC 3339 format

// insertStatement adds one saved PDF to the table.
const insertStatement = "INSERT INTO downloaded_pdfs " + // start of the insert
	"(file_name, pdf_url, source_page, sha256, crc32, size_bytes, downloaded_at) " + // the columns to fill
	"VALUES (?, ?, ?, ?, ?, ?, ?)" // one placeholder per column

// openDatabase opens (or creates) the SQLite file and makes sure the table exists.
func openDatabase(databasePath string) (*sql.DB, error) { // start of function
	database, openError := sql.Open(sqliteDriverName, databasePath) // open the database file
	if openError != nil {                                           // did it work?
		return nil, openError // no, report it
	} // end of check
	database.SetMaxOpenConns(1)                                                           // SQLite allows one writer at a time, so share a single connection
	slog.Debug("opened SQLite database, creating table if missing", "path", databasePath) // log the database setup
	if _, createError := database.Exec(createTableStatement); createError != nil {        // create the table when it is missing
		database.Close()        // release the file
		return nil, createError // report the problem
	} // end of check
	return database, nil // ready to use
} // end of function

// saveRecordToDatabase stores one saved PDF in the database.
func saveRecordToDatabase(database *sql.DB, record DownloadRecord) error { // start of function
	slog.Debug("inserting database row", "file", record.FileName, "sha256", record.SHA256Hex) // log each insert
	_, insertError := database.Exec(insertStatement,                                          // run the insert with these values
		record.FileName, record.PDFURL, record.SourcePage, record.SHA256Hex, record.CRC32Hex, // text columns
		record.SizeInBytes, record.DownloadedAt.Format(time.RFC3339)) // number and timestamp columns
	return insertError // report any problem
} // end of function

// runProgram contains the whole program flow and returns any fatal error.
func runProgram() error { // start of function
	workerCount := min(max(2, runtime.NumCPU()), 8) // use between 2 and 8 workers depending on CPU cores

	runContext, stopListening := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // cancel on Ctrl+C
	defer stopListening()                                                                                  // stop listening when leaving

	pageURLs := slices.Clone(seedPages)                                                                                                                                             // start with the built-in pages
	slog.Debug("configuration", "pages", len(pageURLs), "workers", workerCount, "output_folder", outputFolderName, "database", databaseFilePath, "request_timeout", requestTimeout) // log the settings in use

	if makeFolderError := os.MkdirAll(outputFolderName, 0o755); makeFolderError != nil { // make sure the folder exists
		return makeFolderError // report problems
	} // end of check
	registry := newFileRegistry()                                                     // empty registry
	if indexError := registry.loadExistingPDFs(outputFolderName); indexError != nil { // remember PDFs from earlier runs
		return fmt.Errorf("indexing existing PDFs: %w", indexError) // report problems
	} // end of check

	database, databaseError := openDatabase(databaseFilePath) // open (or create) the SQLite database
	if databaseError != nil {                                 // did it work?
		return fmt.Errorf("opening database %s: %w", databaseFilePath, databaseError) // no, report it
	} // end of check
	defer database.Close() // close the database when leaving

	cookieJar, cookieError := cookiejar.New(nil) // create a cookie store
	if cookieError != nil {                      // did it work?
		return cookieError // no, report it
	} // end of check
	httpClient := &http.Client{ // configure the HTTP client
		Timeout: requestTimeout, // overall time limit per request
		Jar:     cookieJar,      // keep cookies
		Transport: &http.Transport{ // custom connection settings
			TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12}, // require TLS 1.2 or newer
			ForceAttemptHTTP2:   true,                                      // allow HTTP/2
			MaxIdleConnsPerHost: 8,                                         // reuse up to 8 idle connections per host
			IdleConnTimeout:     60 * time.Second,                          // close idle connections after a minute
		}, // end of transport
	} // end of client
	var contentFetcher ContentFetcher = &HTTPContentFetcher{httpClient: httpClient} // wrap the client in our interface

	startTime := time.Now()                                                                    // remember when we started
	jobs := scrapeAllPages(runContext, contentFetcher, pageURLs, workerCount)                  // find every PDF link
	slog.Debug("PDF discovery finished", "unique_pdf_urls", len(jobs), "pages", len(pageURLs)) // report the count

	downloader := &pdfDownloader{contentFetcher: contentFetcher, registry: registry, outputFolder: outputFolderName, queue: list.New(), database: database} // set up downloading
	for _, job := range jobs {                                                                                                                              // queue every job
		downloader.queue.PushBack(job) // add to the queue
	} // end of loop
	downloader.runWorkers(runContext, workerCount)                                                                                                                  // download everything
	slog.Debug("all downloads finished", "saved", downloader.savedCount.Load(), "skipped", downloader.skippedCount.Load(), "failed", downloader.failedCount.Load()) // log the totals

	slices.SortFunc(downloader.records, func(first, second DownloadRecord) int { return cmp.Compare(first.FileName, second.FileName) }) // sort the summary table by name

	tableWriter := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0) // aligned columns on screen
	fmt.Fprintln(tableWriter, "NAME\tBYTES\tSHA256")               // table header
	for _, record := range downloader.records {                    // one row per saved PDF
		fmt.Fprintf(tableWriter, "%s\t%d\t%s…\n", record.FileName, record.SizeInBytes, record.SHA256Hex[:12]) // row with a short hash
	} // end of loop
	tableWriter.Flush()                                   // print the table
	fmt.Printf("\nsaved=%d skipped=%d failed=%d in %s\n", // final summary line
		downloader.savedCount.Load(), downloader.skippedCount.Load(), downloader.failedCount.Load(), time.Since(startTime).Round(time.Millisecond)) // counts and elapsed time

	if runContext.Err() != nil { // was the run interrupted?
		return runContext.Err() // report that
	} // end of check
	return nil // everything went fine
} // end of function

func main() { // program entry point
	debugHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}) // handler that shows everything down to debug level
	slog.SetDefault(slog.New(debugHandler).With("program", "uvpdf"))                                              // use it for the whole program
	slog.Debug("program starting")                                                                                // log the start
	if runError := runProgram(); runError != nil {                                                                // run and check for failure
		slog.Debug("fatal error", "error", runError) // log the failure
		os.Exit(1)                                   // exit with a failure code
	} // end of check
	slog.Debug("program finished successfully") // log the end
} // end of main
