package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	podb "github.com/rehoy/audioplayer/db"
)

type Person struct {
	Age      int
	Height   int
	Hobby    string
	EyeColor string
}

type Server struct {
	Podcast           *podb.Podcast
	TemplateDirectory string
	DB                *podb.DB
}

func NewServer() *Server {
	db := podb.NewDB()

	return &Server{
		Podcast: loadPodcast(),
		DB:      db,
	}
}

func loadPodcast() *podb.Podcast {
	file, err := os.Open("podcasts.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	podcast := &podb.Podcast{}
	err = json.NewDecoder(file).Decode(podcast)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("number of episodes:", len(podcast.Episodes))
	return podcast
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/main/" + "index.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title    string
		Subtitle string
	}{
		Title:    "Podcast Picker",
		Subtitle: "Click the button to get a podcast recommendation",
	}

	err = tmpl.Execute(w, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) clickHandler(w http.ResponseWriter, r *http.Request) {
	person := Person{
		Age:      30,
		Height:   180,
		Hobby:    "Photography",
		EyeColor: "Blue",
	}

	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/" + "infoCard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Button is clicked")
	w.Header().Set("Content-Type", "text/html")

	err = tmpl.Execute(w, person)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) podcastHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/main/" + "podcast.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := r.URL.Query()
	podcast_name := query.Get("name")
	if podcast_name == "" {
		podcast_name = "Underunderstood"
		fmt.Println("no parameter provided")
	}
	fmt.Println("podcast name:", podcast_name)
	episodes, err := s.DB.GetEpisodesFromSeries(podcast_name)
	if err != nil {
		fmt.Println("Error getting episodes from series:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	episodeMap := make(map[string]podb.Episode)
	for _, episode := range episodes {
		episodeMap[episode.Title] = episode
	}

	podcast := podb.Podcast{
		Title:       podcast_name,
		Description: "A podcast about the unknown",
		Episodes:    episodeMap,
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, podcast)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) navbarHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/header/" + "navbar.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/html")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) playerHandler(w http.ResponseWriter, r *http.Request) {
	episode := podb.Episode{
		Episode_id:  0,
		Title:       "unknown",
		Pubdate:     "unknown",
		Description: "unknown",
		AudioURL:    "unknown",
		ImageURL:    "unknown",
	}
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/footer/" + "player.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/html")
	err = tmpl.Execute(w, episode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) idHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	fmt.Println(id)
}

func (s *Server) episodeHandler(w http.ResponseWriter, r *http.Request) {
	id_string := r.URL.Query().Get("id")
	if id_string == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	id, _ := strconv.Atoi(id_string)

	episode := s.DB.GetEpisode(id)

	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/footer/" + "player.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/html")
	err = tmpl.Execute(w, episode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) modalHandler(w http.ResponseWriter, r *http.Request) {
	id_string := r.URL.Query().Get("id")
	var episode podb.Episode
	if id_string == "" {
		episode = podb.Episode{
			Episode_id:  0,
			Title:       "unknown",
			Pubdate:     "unknown",
			Description: "unknown modal description",
			AudioURL:    "unknown",
			ImageURL:    "unknown",
		}
	} else {
		id, err := strconv.Atoi(id_string)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Println(id, "not number")
			return
		}
		episode = s.DB.GetEpisode(id)
	}

	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/main/" + "modal.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/html")
	err = tmpl.Execute(w, episode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) closeModalHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/main/" + "closeModal.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/html")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) faviconHandler(w http.ResponseWriter, r *http.Request) {
	podcast_name := "Underunderstood"
	episodes, err := s.DB.GetEpisodesFromSeries(podcast_name)
	if err != nil {
		fmt.Println("Error getting episodes from series:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	episode := episodes[0]
	imageURL := episode.ImageURL
	fmt.Println(imageURL, len(episodes))
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<link rel="icon" href="%s" type="image/png">`, imageURL)
}

func (s *Server) getEpisode(id int) podb.Episode {
	for _, value := range s.Podcast.Episodes {
		if value.Episode_id == id {
			return value
		}
	}
	return podb.Episode{}
}

func (s *Server) selectorHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/main/" + "podcast-selector.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	titles, err := s.DB.GetSeries()
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Podcast titles", titles)
	podcast_titles := struct {
		PodcastTitles []string
	}{titles}

	w.Header().Set("content-type", "text/html")
	err = tmpl.Execute(w, podcast_titles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) profileHandler(w http.ResponseWriter, r *http.Request){
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/profile/" + "profile.html")
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	series, err := s.DB.GetSeries()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	series_struct := struct{
		Series []string
	}{series}

	w.Header().Set("content-type", "text/html")
	err = tmpl.Execute(w, series_struct)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) SetupServer(folder string) {
	s.TemplateDirectory = folder
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/click", s.clickHandler)
	http.HandleFunc("/podcast", s.podcastHandler)
	http.HandleFunc("/navbar", s.navbarHandler)
	http.HandleFunc("/player", s.playerHandler)
	http.HandleFunc("/id", s.idHandler)
	http.HandleFunc("/episode", s.episodeHandler)
	http.HandleFunc("/modal", s.modalHandler)
	http.HandleFunc("/closemodal", s.closeModalHandler)
	http.HandleFunc("/favicon", s.faviconHandler)
	http.HandleFunc("/podcast-selector", s.selectorHandler)
	http.HandleFunc("/profile", s.profileHandler)
}
