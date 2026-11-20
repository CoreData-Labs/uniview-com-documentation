# 📹 Uniview Documentation Archive

A single, organized, verifiable collection of the PDF documents published on
[uniview.com](https://www.uniview.com/) for selected Uniview IP network camera models,
plus the small Go tool (`main.go`) that collects them and a script (`upload.sh`) that publishes updates. 🗂️

- 📁 **One folder, no duplicates:** every PDF lives in `PDFs/`, and the same document is never stored twice.
- 🔐 **Verifiable:** each file has a SHA-256 fingerprint, so you can prove a copy is unchanged.
- 🧾 **Traceable:** a SQLite database records where every file came from and when it was saved.
- 🔁 **Easy to refresh:** `go run main.go` to update, `./upload.sh` to publish. No arguments needed.
- 📱 **For everyone:** if you can open a PDF on a computer or a phone, you can use this data.

---

## 📚 Table of contents

1. [What is this project for?](#-what-is-this-project-for)
2. [Who benefits? (computer and phone users)](#-who-benefits-computer-and-phone-users)
3. [Covered products](#-covered-products)
4. [Data at a glance (verified specifications)](#-data-at-a-glance-verified-specifications)
5. [Documents found on the product pages](#-documents-found-on-the-product-pages)
6. [Camera terms explained simply](#-camera-terms-explained-simply)
7. [How to use this data](#-how-to-use-this-data-in-simple-terms)
8. [Learn by doing: quick calculators](#-learn-by-doing-quick-calculators)
9. [Update the data (Linux)](#-update-the-data-linux)
10. [Repository layout](#-repository-layout)
11. [How the collector works](#-how-the-collector-works)
12. [Using the database](#-using-the-database)
13. [License and disclaimer](#-license-and-disclaimer)

---

## 🎯 What is this project for?

| Purpose                         | What it means                                                                                                                        |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| 💾 **Backup**                   | Keep your own safe copy of the documents, in case the vendor's website changes, moves a link, or removes a file.                     |
| 🎓 **Education**                | Learn how IP cameras work, how to read a datasheet, and how to plan an installation.                                                 |
| 🤖 **AI training and research** | Use the clean, deduplicated text as input for search tools, chatbots, summarizers and other AI experiments.                          |
| 🛠️ **Building new products**    | Reuse the same approach to build tools like this one: a documentation downloader, a product-comparison site, or an offline help app. |

> ⚠️ **Please read:** the PDFs belong to Uniview. Before using them for AI training or in a commercial product,
> check Uniview's terms of use. See [License and disclaimer](#️-license-and-disclaimer).

---

## 🌍 Who benefits? (computer and phone users)

You do **not** need to be a camera expert or a programmer. Here is what different people can do with this data.

### 📱 If you only have a cell phone

- 📖 **Read specifications anywhere.** Open the repository in your phone's browser (or the GitHub app), tap a PDF, and read it. No computer needed.
- ⬇️ **Save a PDF for offline use.** Download it to your phone's Files app so it works without internet, for example on site in a basement or a remote field.
- 📤 **Share it.** Send a datasheet to a customer, a family member or a technician by WhatsApp, email or any app.
- 🔎 **Search inside a PDF.** Use your PDF viewer's search to jump to words like "PoE", "IP67" or "microSD".
- 🤖 **Ask an AI assistant.** Upload a datasheet to an AI assistant app and ask plain questions such as "Can this camera work outdoors in the rain?".

### 💻 If you have a computer

- 🗂️ **Keep a local library** of all the documents and search them with your computer's file search.
- 🗃️ **Browse the database** with a free viewer (such as DB Browser for SQLite) to see where every file came from.
- 🔄 **Refresh the archive** yourself on Linux with two commands (see [Update the data](#-update-the-data-linux)).
- 🧩 **Build on it.** Extract text, make a search page, or train and test AI tools (subject to Uniview's terms).

### 👥 Examples of who gets value

| Who                             | How they benefit                                                                                                                                                      |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 🏠 **Home owner or shop owner** | Check whether a camera suits your door, yard or shop before you buy: lens angle, night-vision distance, weather rating, and whether it needs a special power adapter. |
| 🔧 **Installer or technician**  | Carry datasheets, quick guides and the user manual in your pocket, even with no signal.                                                                               |
| 🛒 **Reseller or buyer**        | Compare models quickly and give customers accurate specifications.                                                                                                    |
| 🛡️ **IT or security manager**   | Check supported security features (HTTPS, user levels, ONVIF) and keep proof of what the documents said.                                                              |
| 🎓 **Student or teacher**       | Learn real-world camera terms and practice reading professional datasheets.                                                                                           |
| 🤖 **AI or data hobbyist**      | Practice text extraction, search and question answering on real technical documents.                                                                                  |
| 🧑‍💻 **Developer or startup**     | Use the collector as a starting point for your own documentation or product tools.                                                                                    |

---

## 📦 Covered products

The collector visits these seven Uniview product pages and downloads every PDF they link to.

| Model                 | Series           | Product page                                                                                               |
| --------------------- | ---------------- | ---------------------------------------------------------------------------------------------------------- |
| IPC2122LB-ASF28K-A    | Easy / EasyBasic | [page](https://www.uniview.com/Products/Network_Cameras/Easy_Series/EasyBasic_Series/IPC2122LB-ASF28K-A/)  |
| IPC2108SB-ADF16KM-I0  | Prime / PRIME-I  | [page](https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2108SB-ADF16KM-I0/)  |
| IPC2225SB-ADF40KM-I1  | Prime / PRIME-I  | [page](https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2225SB-ADF40KM-I1/)  |
| IPC2222SB-AHDF40KM-I1 | Prime / PRIME-I  | [page](https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2222SB-AHDF40KM-I1/) |
| IPC2222SB-AHDF28KM-I1 | Prime / PRIME-I  | [page](https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2222SB-AHDF28KM-I1/) |
| IPC2225SB-ADF28KM-I1  | Prime / PRIME-I  | [page](https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2225SB-ADF28KM-I1/)  |
| IPC2322SB-HDZK-PI-I0  | Prime / PRIME-I  | [page](https://www.uniview.com/Products/Network_Cameras/Prime_Series/PRIMEI_Series/IPC2322SB-HDZK-PI-I0/)  |

---

## 📊 Data at a glance (verified specifications)

The values below were read from the official Uniview product pages on **2026-11-20** for two of the seven models.
Always check the datasheet for the final word, because Uniview can change specifications.

| Specification                    | IPC2122LB-ASF28K-A                      | IPC2108SB-ADF16KM-I0                                                                                                          |
| -------------------------------- | --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| **Type**                         | 2MP fixed IR bullet network camera      | 8MP OmniView fixed IR bullet network camera                                                                                   |
| **Max resolution**               | 2 MP (1920 × 1080)                      | 8 MP (3840 × 2160)                                                                                                            |
| **Sensor**                       | 1/2.9" CMOS                             | 1/2.7" CMOS                                                                                                                   |
| **Lens (focal length)**          | 2.8 mm fixed                            | 1.68 mm fixed                                                                                                                 |
| **Field of view (H / V / D)**    | 101.1° / 56.5° / 117°                   | 180° / 100° / 180°                                                                                                            |
| **Min. illumination**            | 0.01 lux color (F2.0), 0 lux with IR on | 0.01 lux color (F2.0), 0 lux with IR on                                                                                       |
| **Night vision (IR distance)**   | up to 30 m (98.4 ft)                    | up to 20 m (65.6 ft)                                                                                                          |
| **WDR (bright and dark scenes)** | DWDR                                    | 120 dB                                                                                                                        |
| **Video compression**            | Ultra 265, H.265, H.264                 | Ultra 265, H.265, H.264, MJPEG                                                                                                |
| **Max frame rate**               | 30 fps (main stream)                    | 30 fps (main stream)                                                                                                          |
| **Bit rate range**               | 128 Kbps to 6144 Kbps                   | 128 Kbps to 16 Mbps                                                                                                           |
| **Streams**                      | Dual                                    | Triple                                                                                                                        |
| **Smart features**               | Motion and audio detection              | Smart intrusion prevention (cross line, intrusion, enter/leave area) with human, vehicle and non-motor-vehicle classification |
| **Built-in microphone**          | Yes                                     | Yes                                                                                                                           |
| **Edge storage**                 | microSD up to 512 GB                    | microSD up to 1 TB                                                                                                            |
| **Network**                      | 1 × RJ45 10/100 Mbps                    | 1 × RJ45 10/100 Mbps                                                                                                          |
| **Integration**                  | ONVIF (S, G, T), API, SDK               | ONVIF (S, G, T, M), API, SDK                                                                                                  |
| **Power**                        | DC 12 V, PoE (IEEE 802.3af)             | DC 12 V, PoE (IEEE 802.3af)                                                                                                   |
| **Max power use**                | 5.0 W                                   | 5.5 W                                                                                                                         |
| **Weather protection**           | IP67                                    | IP67                                                                                                                          |
| **Working temperature**          | -30 °C to 60 °C (-22 °F to 140 °F)      | -30 °C to 60 °C (-22 °F to 140 °F)                                                                                            |
| **Surge protection**             | 4 kV                                    | 4 kV                                                                                                                          |
| **Size (L × W × H)**             | 161 × 62 × 63 mm                        | 198 × 71 × 71 mm                                                                                                              |
| **Weight**                       | 0.21 kg (0.46 lb)                       | 0.42 kg (0.93 lb)                                                                                                             |
| **Body material**                | Graphene + plastic                      | Metal                                                                                                                         |
| **Web interface languages**      | 1 (English)                             | 20 languages                                                                                                                  |
| **Certifications listed**        | CE (EMC, LVD, RoHS), FCC, WEEE          | CE (EMC, LVD, RoHS), FCC, UL/CUL, WEEE                                                                                        |

### 📏 How far can each camera "see"? (DORI distances)

DORI is an international way to describe what a camera can do at a given distance. Distances are from the datasheets.

| DORI level       | What it means                                              | IPC2122LB-ASF28K-A (2.8 mm) | IPC2108SB-ADF16KM-I0 (1.68 mm) |
| ---------------- | ---------------------------------------------------------- | --------------------------- | ------------------------------ |
| 👁️ **Detect**    | You can tell something is there                            | 45.0 m (147.6 ft)           | 37.8 m (124.0 ft)              |
| 🔭 **Observe**   | You can see its general features, like a person's clothing | 18.0 m (59.1 ft)            | 15.1 m (49.6 ft)               |
| 🧑 **Recognize** | You can recognize a known person                           | 9.0 m (29.5 ft)             | 7.6 m (24.9 ft)                |
| 🪪 **Identify**   | You can identify a person beyond reasonable doubt          | 4.5 m (14.8 ft)             | 3.8 m (12.5 ft)                |

> 💡 **Takeaway:** the 8 MP OmniView model covers a very wide area (180°), while the 2 MP model sees a little farther in a narrower view.
> Choose by what you need to watch, not only by megapixels.

> 📝 The other five models are listed on their product pages above. This README does not yet include their verified numbers.

---

## 📄 Documents found on the product pages

On the pages checked, each model offers these downloads. Titles and dates are as shown by Uniview on 2026-11-20.

| Type           | Title (as shown on the site)                                                              | Language | Date shown | Model-specific?                             |
| -------------- | ----------------------------------------------------------------------------------------- | -------- | ---------- | ------------------------------------------- |
| 📘 Datasheet   | UNV IPC2122LB-ASF28(40)K-A 2MP Fixed IR Bullet Network Camera.pdf                         | English  | 2026-06-03 | Yes (covers the 2.8 mm and 4.0 mm versions) |
| 📘 Datasheet   | UNV IPC2108SB-ADF16KM-I0 8MP HD Wide Angle Intelligent IR Fixed Bullet Network Camera.pdf | English  | 2026-10-19 | Yes                                         |
| 📗 User manual | UNV-Network Camera User Manual.pdf                                                        | English  | 2025-12-10 | Shared by both pages                        |
| 📙 Quick guide | Quick Guide.pdf                                                                           | English  | 2026-07-03 | Shared by both pages                        |

Good to know:

- 🔁 The user manual and the quick guide use the **same download link** on both pages, so they are saved **once**. This is the duplicate protection at work.
- 🖼️ Each page also offers PNG images (product photos and dimension drawings). The collector downloads **PDFs only**.
- 🧮 Expect roughly **one datasheet per model plus the shared manual and quick guide**. The exact total is whatever is in the `downloaded_pdfs` table.
- 🏷️ **About file names:** Uniview hosts these PDFs under ID-style addresses (long random-looking names), so saved files may have ID-style names.
  To find which model a file belongs to, look it up in the database (`source_page` column) instead of relying on the file name.

---

## 🔤 Camera terms explained simply

| Term                               | Simple meaning                                                                                                                      |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| **IP camera / network camera**     | A camera that sends video over a network cable or Wi-Fi, like a tiny computer.                                                      |
| **Bullet camera**                  | A tube-shaped camera, usually mounted on a wall or pole.                                                                            |
| **MP (megapixel)**                 | How many million dots make up the picture. More MP means more detail, but also bigger video files.                                  |
| **1080P / 8 MP (4K)**              | 1080P is 1920 × 1080 pixels. 8 MP is 3840 × 2160 pixels, also called 4K.                                                            |
| **Sensor size (1/2.9", 1/2.7")**   | The size of the chip that captures light. A bigger sensor generally helps in dim light.                                             |
| **Focal length (2.8 mm, 1.68 mm)** | A smaller number gives a wider view. A larger number zooms in on distant things.                                                    |
| **Field of view (FOV)**            | How wide the camera sees, in degrees. 180° is like looking straight along a whole wall.                                             |
| **IR / Smart IR**                  | Invisible infrared lights that let the camera see in the dark. "Smart" means the lights adjust so close objects are not washed out. |
| **Lux**                            | A measure of light. 0.01 lux is very dark, like a moonless night with a little glow.                                                |
| **WDR / DWDR**                     | Helps show both very bright areas (a window) and dark areas in the same picture.                                                    |
| **H.264 / H.265 / Ultra 265**      | Ways to compress video. Newer ones (H.265, Ultra 265) make smaller files at similar quality.                                        |
| **Bit rate (Kbps / Mbps)**         | How much data the video uses per second. Higher means better quality and bigger files.                                              |
| **fps**                            | Frames per second. 30 fps looks smooth.                                                                                             |
| **PoE (IEEE 802.3af)**             | Power over Ethernet: one network cable carries both data and power, so you need no separate power adapter.                          |
| **DC 12 V**                        | The alternative: a 12-volt power adapter.                                                                                           |
| **IP67**                           | Dust-tight and protected against temporary immersion in water. Suitable for outdoor use.                                            |
| **microSD (edge storage)**         | A memory card inside the camera that keeps recording even if the network goes down.                                                 |
| **ONVIF**                          | A common standard that lets cameras from different brands work with the same recorders and apps.                                    |
| **API / SDK**                      | Tools for programmers to connect their own software to the camera.                                                                  |
| **Surge protection (4 kV)**        | Protection against power spikes, such as those from lightning.                                                                      |
| **NAS / ANR**                      | Network storage options. ANR helps the camera fill gaps in recordings after a network problem.                                      |
| **EZStation / APP**                | Uniview's computer software and phone app used to view and manage cameras.                                                          |

---

## 🚀 How to use this data (in simple terms)

You do not need to be a programmer for the first four options.

### 1️⃣ Just read the documents 📖

1. Open the **`PDFs/`** folder (on GitHub, on your computer, or on your phone).
2. Open a PDF with your normal viewer.
3. Use the viewer's search to find a word such as `IP67` or `PoE`.

### 2️⃣ Find out which model a file belongs to 🕵️

1. Install a free viewer such as [DB Browser for SQLite](https://sqlitebrowser.org/).
2. Open **`uvpdf.db`** and then the table **`downloaded_pdfs`**.
3. Each row shows the file name, its web address, the product page it came from, and when it was saved.

### 3️⃣ Check that a file is genuine ✅

1. Compute the file's fingerprint: `sha256sum PDFs/<file-name>.pdf`
2. Look up the stored fingerprint: `sqlite3 uvpdf.db "SELECT sha256 FROM downloaded_pdfs WHERE file_name = '<file-name>.pdf';"`
3. If the two values match, the file is exactly the same as when it was downloaded.

### 4️⃣ Use it for learning 🎓

- Read a datasheet next to the [spec table above](#-data-at-a-glance-verified-specifications) and match each term to the glossary.
- Compare two models and write down the differences (view angle, night distance, storage, smart features).
- Practice a complete installation plan on paper: where to mount, how to power it, how to store the video.

### 5️⃣ Use it for AI projects 🤖

1. Extract the text from each PDF (for example with `pdftotext` or a Python PDF library).
2. Split the text into small chunks and keep the file name and source page with each chunk.
3. Feed the chunks into your search tool, chatbot or fine-tuning experiment.
4. Because every file has a hash and a source URL, you can always trace an answer back to its document.

> ⚠️ Check Uniview's terms before training or publishing a model with this content.

### 6️⃣ Build your own product 🛠️

- 📥 A **documentation downloader** for any vendor: change the page list in `main.go`.
- 🔍 A **searchable help site** for camera models.
- 📊 A **comparison tool** that reads the database and the PDFs.
- 📱 An **offline field guide** app for installers.

The code is small and commented on every line, so it is a good starting point to learn from and extend. 🌱

---

## 🧮 Learn by doing: quick calculators

These are simple estimates for learning. Real results depend on the scene, the camera settings and the card.

### 💽 How long will a microSD card record?

```
hours of recording = card size in GB  ÷  (bit rate in Mbps × 0.45)
```

(One Mbps running for one hour is about 0.45 GB.)

Example: a 512 GB card with a video bit rate of 4 Mbps:
`512 ÷ (4 × 0.45) = 512 ÷ 1.8 ≈ 284 hours ≈ 11.8 days` of nonstop recording.

Tips: lower bit rate, lower resolution or motion-only recording all make the card last longer.
Both verified models list a bit rate range starting at 128 Kbps.

### 🔌 Power: PoE or adapter?

- **PoE:** plug one network cable into a PoE switch or injector that supports **IEEE 802.3af**. Both verified models support it.
- **Adapter:** use a 12 V DC supply with the 5.5 mm coaxial plug listed in the datasheet.
- Both verified models draw 5.5 W or less, which is well within what a standard PoE port provides.

### 🧭 Which camera for which place?

| Your situation                                        | Look at this in the datasheet                                        |
| ----------------------------------------------------- | -------------------------------------------------------------------- |
| Watch a long driveway or fence line                   | Narrower field of view (larger focal length) and long DORI distances |
| Cover a wide yard, entrance or corner with one camera | Wide field of view (for example 180°)                                |
| Dark area with no lights                              | IR distance, minimum illumination, WDR                               |
| Outdoors in rain, dust or cold                        | IP rating (IP67) and working temperature range                       |
| Need to keep recording if the network fails           | microSD card support and maximum size                                |
| Want to connect to other brands' recorders            | ONVIF profiles, API/SDK                                              |

---

## 🐧 Update the data (Linux)

To refresh the archive you only need **Go** and **Git** on a Linux computer. The steps are the same every time.

### Step 1️⃣ Install Go (one time)

Pick the line for your Linux:

```bash
# Ubuntu / Debian / Linux Mint
sudo apt update && sudo apt install -y golang-go git

# Fedora
sudo dnf install -y golang git

# Arch / Manjaro
sudo pacman -S go git
```

Check it worked (you need Go 1.22 or newer):

```bash
go version
```

If your Linux ships an older Go, install the newest from [go.dev/dl](https://go.dev/dl/):

```bash
# replace <VERSION> with the file name shown on go.dev/dl, for example go1.xx.x.linux-amd64.tar.gz
curl -LO https://go.dev/dl/<VERSION>
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf <VERSION>
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile && source ~/.profile
go version
```

### Step 2️⃣ Get the repository and set up Git (one time)

```bash
git clone https://github.com/PrajwalKoirala638/uniview-com-documentation.git
cd uniview-com-documentation

# tell Git who you are (used on the commits)
git config user.name  "Your Name"
git config user.email "you@example.com"

# make the upload script runnable
chmod +x upload.sh

# download the SQLite driver used by main.go
go mod tidy
```

> 🔑 **Permission to upload:** to push to GitHub you must be allowed to write to the repository
> and signed in. The simplest ways are a GitHub **personal access token** (used as the password when Git asks)
> or an **SSH key**. Never paste a token or password into any file in this repository.

### Step 3️⃣ Update the data 🔄

```bash
go run main.go
```

What happens:

- the program visits the seven product pages,
- downloads **only the PDFs it has not seen before**,
- saves them in `PDFs/`,
- and adds a row for each new file to `uvpdf.db`.

Want the details? Everything is logged at debug level on screen. To keep a copy:

```bash
go run main.go 2> uvpdf.log
```

### Step 4️⃣ Upload the new data ⬆️

```bash
./upload.sh
```

What `upload.sh` does:

1. checks that Git is installed and that you are inside the repository,
2. stages everything that changed (new PDFs, `uvpdf.db`, code),
3. says "Nothing new to upload" if there are no changes,
4. otherwise commits with a timestamped message and pushes to the current branch.

Use your own message if you like:

```bash
./upload.sh "Add new datasheets for November"
```

### ✅ The whole update in 3 lines

```bash
cd uniview-com-documentation
go run main.go
./upload.sh
```

### 🧯 If something goes wrong

| Problem                              | What to do                                                            |
| ------------------------------------ | --------------------------------------------------------------------- |
| `go: command not found`              | Install Go (Step 1) and open a new terminal.                          |
| `go.mod` or driver errors            | Run `go mod tidy` again, with internet access.                        |
| `Permission denied` on `./upload.sh` | Run `chmod +x upload.sh`.                                             |
| Git asks for a password              | Use a personal access token, or set up an SSH key.                    |
| `push` rejected                      | Run `git pull --rebase` and then `./upload.sh` again.                 |
| Some pages report errors             | Check the log. The site may be slow or have changed. Run again later. |

---

## 🗺️ Repository layout

```
.
├── PDFs/              # every downloaded PDF, in one flat folder
├── uvpdf.db           # SQLite database with one row per saved PDF
├── main.go            # the collector (Go)
├── go.mod             # Go module file
├── upload.sh          # commits and pushes new data to GitHub
├── LICENSE            # MIT license
└── README.md
```

---

## ⚙️ How the collector works

1. 🌐 Visits the seven product pages listed above.
2. 🔗 Finds every link that points to a `.pdf`, in page attributes and in script data.
3. 🚫 Skips any file whose **name** was already saved, including PDFs left by earlier runs.
4. ⬇️ Downloads the rest and checks that each really starts with the `%PDF` signature.
5. 🔐 Skips any file whose **SHA-256** matches one already saved.
6. 💾 Saves new files into `PDFs/` and records each in `uvpdf.db`.

**File-name rules:** names never start or end with `_`, never contain `__`, and always end in `.pdf`.

Also:

- 🔄 Failed requests are retried with backoff.
- 🧱 Files are written under a temporary name and renamed when complete, so a half-written PDF never appears.
- 🪵 Everything is logged at debug level to standard error.

**Add more products:** add the product page address to the `seedPages` list at the top of `main.go` and run `go run main.go` again.

---

## 🗃️ Using the database

Open it with the `sqlite3` command-line tool (or any SQLite viewer):

```bash
sqlite3 uvpdf.db
```

Columns of the `downloaded_pdfs` table: `file_name`, `pdf_url`, `source_page`, `sha256`, `crc32`, `size_bytes`, `downloaded_at`.

```sql
-- every saved document, largest first
SELECT file_name, size_bytes FROM downloaded_pdfs ORDER BY size_bytes DESC;

-- which product page each document came from
SELECT file_name, source_page FROM downloaded_pdfs ORDER BY source_page, file_name;

-- how many documents each product page contributed
SELECT source_page, COUNT(*) AS documents FROM downloaded_pdfs GROUP BY source_page;

-- what was added most recently
SELECT file_name, downloaded_at FROM downloaded_pdfs ORDER BY downloaded_at DESC LIMIT 10;
```

---

## ⚖️ License and disclaimer

- 📄 The code and repository contents authored here are released under the [MIT License](LICENSE).
- 🏢 The PDF documents are **owned by Uniview** and remain subject to their own terms and copyrights.
  The MIT license does not apply to them. They are collected from publicly available pages on uniview.com.
- 🤝 This project is **not affiliated with or endorsed by Uniview**.
- 🔒 Before using the documents for **AI training or in a commercial product**, check Uniview's terms of use
  or ask Uniview for permission.
- 🕰️ Specifications and documents change. The numbers in this README were read on 2026-11-20. Always check
  [uniview.com](https://www.uniview.com/) or the datasheet for the latest values before relying on them.
- 🛠️ The calculators and tips are simple learning aids, not professional design advice.
