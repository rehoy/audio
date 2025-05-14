package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"io"

	podb "github.com/rehoy/audioplayer/db"
)

type Person struct {
	Age      int
	Height   int
	Hobby    string
	EyeColor string
}

type Server struct {
	TemplateDirectory string
	DB                *podb.DB
}

func NewServer() *Server {
	db := podb.NewDB()

	return &Server{
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
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/main/" + "podcast-subcontainer.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse form values from the request
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	podcast_name := r.FormValue("name")
	orderBy := r.FormValue("ordering")
	upDown := r.FormValue("up-down")


	if podcast_name == "" {
		podcast_name = "Underundertood"
		fmt.Println("no parameter provided")
	}

	var ordering string
	var sorting string

	switch orderBy {
	case "title":
		ordering = "title"
	case "pubdate":
		ordering = "pubdate"
	}

	switch upDown {
	case "up":
		sorting = "ASC"
	case "down":
		sorting = "DESC"
	}
	episodes, err := s.DB.GetEpisodesFromSeries(podcast_name, ordering, sorting)
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
		Episodes:    episodes,
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

	data := struct {
		Title       string
		Description template.HTML
	}{
		Title:       episode.Title,
		Description: template.HTML(episode.Description),
	}

	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/main/" + "modal.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/html")
	err = tmpl.Execute(w, data)
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

func (s *Server) profileHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(s.TemplateDirectory + "/profile/" + "profile.html")
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

func (s *Server) renderPodcastOverview(w http.ResponseWriter, tmpl *template.Template) {
	series, err := s.DB.GetSeries()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	series_struct := struct {
		Series []string
	}{series}

	fmt.Println("Series:", series)

	w.Header().Set("content-type", "text/html")
	err = tmpl.Execute(w, series_struct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) overviewHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/profile/podcast-overview.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.renderPodcastOverview(w, tmpl)
	case http.MethodDelete:
		query := r.URL.Query()
		name := query.Get("name")
		if name == "" {
			http.Error(w, "Missing name", http.StatusBadRequest)
			return
		}

		series_id, err := s.DB.GetSeriesIDByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println("Error getting series ID by name:", err)
			return
		}

		err = s.DB.DeleteSeries(series_id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println("Error deleting series ID by name:", err)
			return
		}

		s.renderPodcastOverview(w, tmpl)
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "Missing name", http.StatusBadRequest)
			return
		}
		
		fmt.Println("Adding podcastURL:", name)
		err = s.DB.AddPodcastToDB(name, "pod.db")
		if err != nil {
			fmt.Println("Error adding podcast to DB:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		s.renderPodcastOverview(w, tmpl)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (db *Server)podcastContainerHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/main/podcast.html")
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

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
    // Get the file URL from the request
    fileURL := r.URL.Query().Get("url")
    if fileURL == "" {
        http.Error(w, "Missing file URL", http.StatusBadRequest)
        return
    }

    // Extract the filename from the URL
    filename := filepath.Base(fileURL)

    // Download the file from the URL
    resp, err := http.Get(fileURL)
    if err != nil {
        http.Error(w, "Error downloading file", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    // Set the appropriate headers for file download
    w.Header().Set("Content-Disposition", "attachment; filename="+filename)
    w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
    w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))

    // Copy the file content to the response writer
    _, err = io.Copy(w, resp.Body)
    if err != nil {
        http.Error(w, "Error copying file to response", http.StatusInternalServerError)
        return
    }
}





func (s *Server) SetupServer(folder string) {
	s.TemplateDirectory = folder
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/click", s.clickHandler)
	http.HandleFunc("/podcast", s.podcastContainerHandler)
	http.HandleFunc("/navbar", s.navbarHandler)
	http.HandleFunc("/player", s.playerHandler)
	http.HandleFunc("/id", s.idHandler)
	http.HandleFunc("/episode", s.episodeHandler)
	http.HandleFunc("/modal", s.modalHandler)
	http.HandleFunc("/closemodal", s.closeModalHandler)
	http.HandleFunc("/favicon", s.faviconHandler)
	http.HandleFunc("/podcast-selector", s.selectorHandler)
	http.HandleFunc("/profile", s.profileHandler)
	http.HandleFunc("/podcast-overview", s.overviewHandler)
	http.HandleFunc("/podcast-container", s.podcastHandler)


}
