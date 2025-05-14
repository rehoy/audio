package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/signal"

// 	"github.com/rehoy/audioplayer/server"
// )

// func main() {
// 	s := server.NewServer()
// 	s.SetupServer("templates")

// 	http.Handle("/style.css", http.FileServer(http.Dir(".")))

// 	fmt.Println("listening on http://localhost:8080/")

// 	go func() {
// 		err := http.ListenAndServe(":8080", nil)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}();

// 	c := make(chan os.Signal, 1)
// 	stop := make(chan bool)

// 	go func() {
// 		<-c
// 		stop <- true
// 	}()

// 	signal.Notify(c, os.Interrupt)
// 	<-c
// 	log.Println("Shutting down...")
// 	s.DB.Close()

// }

import (
	"fmt"
	"github.com/rehoy/audioplayer/db"
)

func main() {

	db := podb.NewDB()

	defer db.Close()
	fmt.Println("Database opened")

	for i := 0; i < 10; i++ {
		series_name, err := db.SeriesNameFromID(i)
		if err != nil {
			continue
		}

		fmt.Println("Series Name: ", series_name)
	}

	episodes, err := db.GetEpisodesFromSeries("Lemonade Stand")
	if err != nil {
		fmt.Println("Error getting episodes: ", err)
		return
	}

	for _, episode := range episodes {
		fmt.Println("Episode: ", episode.Title, " ", episode.Episode_id)
	}
	
	podcast, err := db.GetPodcast(2)
	if err != nil {
		fmt.Println("Error getting podcast: ", err)
		return
	}

	fmt.Println(podcast.Title, podcast.RssFeed)

	podcast_url := "https://anchor.fm/s/101ec0f34/podcast/rss"

	// updated, err := db.podcastFromFeed(podcast_url)

	newEpisodes, err := db.FindNewEpisodes(podcast_url)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, episode := range newEpisodes {
		fmt.Println("title:", episode.Title)
	}


	




}
