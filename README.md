# Concurrent File Downloader (Go)

A simple concurrent file downloader written in Go. It downloads a file using multiple HTTP range requests in parallel, merges the chunks, and shows a real-time progress bar.

---

## 🛠 Features

- Concurrent HTTP downloads using goroutines
- Uses `Range` header for chunked requests
- Real-time progress bar (powered by [`schollz/progressbar`](https://github.com/schollz/progressbar))
- Auto-detects output file extension from `Content-Type`
- Displays total download time

---

## 📦 Requirements

- Go 1.18 or later
- Internet access
- Server must support `Accept-Ranges: bytes`

Install dependencies:

```bash
go get github.com/schollz/progressbar/v3

```
