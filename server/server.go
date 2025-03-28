package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Person struct {
	Age      int
	Height   int
	Hobby    string
	EyeColor string
}

type Podcast struct {
	Title       string
	Description string
	Episodes    map[string]Episode
}

type Episode struct {
	ID          int
	Title       string
	Pubdate     string
	Description string
	AudioURL    string
	ImageURL    string
}

type Server struct {
	Podcast *Podcast
	TemplateDirectory string
}

func NewServer() *Server {
	return &Server{
		Podcast: loadPodcast(),
	}
}

func loadPodcast() *Podcast {
	file, err := os.Open("pod.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	podcast := &Podcast{}
	err = json.NewDecoder(file).Decode(podcast)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("number of episodes:", len(podcast.Episodes))
	return podcast
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/" + "index.html")

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
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/" + "podcast.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, s.Podcast)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) navbarHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/" + "navbar.html")
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
	episode := Episode{
		ID:          0,
		Title:       "unknown",
		Pubdate:     "unknown",
		Description: "unknown",
		AudioURL:    "unknown",
		ImageURL:    "unknown",
	}
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/" + "player.html")
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

	var episode Episode

	for _, value := range s.Podcast.Episodes {
		if value.ID == id {
			episode = value
			break
		}
	}

	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/" + "player.html")
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
	var episode Episode
	if id_string == "" {
		episode = Episode{
			ID:          0,
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
		episode = s.getEpisode(id)
	}

	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/" + "modal.html")
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
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/" + "closeModal.html")
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

func (s *Server) getEpisode(id int) Episode {
	for _, value := range s.Podcast.Episodes {
		if value.ID == id {
			return value
		}
	}
	return Episode{}
}

func(s *Server) SetupServer(folder string) {
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
}