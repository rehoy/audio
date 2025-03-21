package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mmcdole/gofeed"
)

type Podcast struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Episodes    map[string]Episode `json:"episodes"`
}

type Episode struct {
	Title       string `json:"title"`
	PubDate     string `json:"pubDate"`
	Description string `json:"description"`
	AudioURL    string `json:"audioUrl"`
	ImageURL    string `json:"imageUrl"`
}

func WritePodcastsToJSON(feedURL string) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	// Use the itunes:image tag from the channel as the default image
	defaultImageURL := ""
	if feed.ITunesExt != nil && feed.ITunesExt.Image != "" {
		defaultImageURL = feed.ITunesExt.Image
	}

	podcast := Podcast{
		Title:       feed.Title,
		Description: feed.Description,
		Episodes:    make(map[string]Episode),
	}

	for _, item := range feed.Items {
		imageURL := defaultImageURL
		if item.ITunesExt != nil && item.ITunesExt.Image != "" {
			imageURL = item.ITunesExt.Image
		}

		episode := Episode{
			Title:       item.Title,
			PubDate:     item.Published,
			Description: item.Description,
			AudioURL:    item.Enclosures[0].URL,
			ImageURL:    imageURL,
		}
		podcast.Episodes[episode.Title] = episode
	}

	// Read existing data from podcasts.json
	existingData := make(map[string]Podcast)
	file, err := os.Open("podcasts.json")
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&existingData); err != nil {
			return fmt.Errorf("failed to decode existing JSON: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to open JSON file: %w", err)
	}

	// Append the new podcast data
	existingData[podcast.Title] = podcast

	// Write the updated data back to podcasts.json
	file, err = os.Create("podcasts.json")
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(existingData); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

func main() {
	feedURL := flag.String("feed", "", "The RSS feed URL to parse")

	flag.Parse()

	if *feedURL == "" {
		fmt.Println("Error: feed URL is required")
		flag.Usage()
		return
	}

	if err := WritePodcastsToJSON(*feedURL); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Podcast data written to podcasts.json")
}
