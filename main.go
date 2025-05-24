package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

func main() {
	// Define flags
	url := flag.String("url", "", "The URL of the file to download (required)")
	output := flag.String("output", "", "The name of the output file")
	numChunks := flag.Int("chunks", 6, "The number of concurrent download chunks")

	start := time.Now()
	defer func() {
		fmt.Println("⏱ Total time:", time.Since(start))
	}()

	flag.Parse()

	if *url == "" {
		fmt.Println("❌ You must specify a URL with --url")
		flag.Usage()
		os.Exit(1)
	}

	// HEAD request to get metadata
	resp, err := http.Head(*url)
	if err != nil {
		fmt.Println("Request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server responded with status: %s\n", resp.Status)
		os.Exit(1)
	}

	contentLengthStr := resp.Header.Get("Content-Length")
	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		fmt.Println("Invalid Content-Length header")
		os.Exit(1)
	}

	if resp.Header.Get("Accept-Ranges") != "bytes" {
		fmt.Println("❌ Server does not support range requests. Cannot download in chunks.")
		os.Exit(1)
	}

	// Determine file extension or use fallback
	contentType := resp.Header.Get("Content-Type")
	ext := "bin"
	if parts := strings.Split(contentType, "/"); len(parts) == 2 {
		ext = parts[1]
	}

	// Use default output file name if not provided
	outFile := *output
	if outFile == "" {
		outFile = "download." + ext
	}

	// Create progress bar
	bar := progressbar.NewOptions(contentLength,
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(60),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetDescription("Downloading"))

	// Download in chunks
	chunkSize := contentLength / *numChunks
	client := &http.Client{}
	chunks := make([][]byte, *numChunks)
	var wg sync.WaitGroup

	for i := range *numChunks {
		start := i * chunkSize
		end := start + chunkSize - 1
		if i == *numChunks-1 {
			end = contentLength - 1
		}

		wg.Add(1)
		go func(i, start, end int) {
			defer wg.Done()

			req, _ := http.NewRequest("GET", *url, nil)
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
			res, err := client.Do(req)
			if err != nil {
				fmt.Printf("Chunk %d error: %v\n", i, err)
				return
			}
			defer res.Body.Close()

			// TeeReader updates progress bar as we read
			reader := io.TeeReader(res.Body, bar)
			data, err := io.ReadAll(reader)
			if err != nil {
				fmt.Printf("Chunk %d read error: %v\n", i, err)
				return
			}

			chunks[i] = data
		}(i, start, end)
	}

	wg.Wait()

	// Write all chunks to file
	f, err := os.Create(outFile)
	if err != nil {
		fmt.Println("❌ Failed to create file:", err)
		os.Exit(1)
	}
	defer f.Close()

	for _, chunk := range chunks {
		f.Write(chunk)
	}

	fmt.Println("\n✅ Download complete:", outFile)
}
