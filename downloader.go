
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/mmcdole/gofeed"
)

func downloadPodcast(rssURL string, start, end int) {
	// Parse the RSS feed
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(rssURL)
	if err != nil {
		fmt.Println("Failed to parse RSS feed:", err)
		return
	}

	if len(feed.Items) == 0 {
		fmt.Println("No episodes found in the RSS feed.")
		return
	}

	// Create a directory for the podcast
	podcastTitle := feed.Title
	if podcastTitle == "" {
		podcastTitle = "podcast"
	}
	os.MkdirAll(podcastTitle, os.ModePerm)

	// Reverse the order of episodes
	episodes := feed.Items

	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].PublishedParsed.Before(*episodes[j].PublishedParsed)
	})

	// Apply start and end range
	if start > 0 {
		start = start - 1 // Convert to zero-based index
	}
	if end == 0 || end > len(episodes) {
		end = len(episodes)
	}
	episodes = episodes[start:end]

	// Download each episode
	for _, episode := range episodes {
		if len(episode.Enclosures) > 0 {
			for _, enclosure := range episode.Enclosures {
				if enclosure.Type == "audio/mpeg" {
					fileURL := enclosure.URL
					fileName := filepath.Join(podcastTitle, episode.Title+".mp3")
					fmt.Println("Downloading:", episode.Title)
					err := downloadFile(fileURL, fileName)
					if err != nil {
						fmt.Printf("Failed to download %s: %v\n", episode.Title, err)
					} else {
						fmt.Println("Saved:", fileName)
					}
					break
				}
			}
		}
	}
}

func downloadFile(url, fileName string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	rssURL := flag.String("rssurl", "", "The RSS feed URL of the podcast.")
	start := flag.String("start", "0", "The starting episode number (1-based index).")
	end := flag.String("end", "0", "The ending episode number (1-based index).")
	flag.Parse()

	if *rssURL == "" {
		fmt.Println("Error: --rssurl is required.")
		flag.Usage()
		return
	}

	startInt, _ := strconv.Atoi(*start)
	endInt, _ := strconv.Atoi(*end)

	downloadPodcast(*rssURL, startInt, endInt)
}


