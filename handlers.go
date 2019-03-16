package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	log "github.com/Sirupsen/logrus"

	omdb "github.com/kenshaw/imdb"
	goodreads "github.com/prfalken/watchreadlisten/goodreads"
	imdb "github.com/prfalken/watchreadlisten/imdb"
	spotify "github.com/prfalken/watchreadlisten/spotify"
)

type Entry struct {
	ID       string
	Title    string
	Link     string
	ImageURL url.URL
	Type     string
}

// Search Imdb, Goodreads, and Spotify.
func Search(q string, imdbClient imdb.Imdb, grClient goodreads.Goodreads, spClient spotify.Spotify) (i []*omdb.MovieResult, g goodreads.GoodreadsResponse, s spotify.SearchAlbumsResponse) {
	var wg sync.WaitGroup
	wg.Add(3)
	go func(q string) {
		defer wg.Done()
		movies, err := imdbClient.SearchMovies(q)
		if err != nil {
			log.Println("ERROR (imdb SearchMovies):", err.Error())
		}
		for _, mov := range movies {
			i = append(i, mov)
		}
	}(q)
	go func(q string) {
		defer wg.Done()
		books, err := grClient.SearchBooks(q)
		if err != nil {
			log.Println("ERROR (gr SearchBooks):", err.Error())
		}
		for i, w := range books.Search.Works {
			w.BestBook.Title = truncate(w.BestBook.Title, "...", 60)
			books.Search.Works[i] = w
		}
		g = books
	}(q)
	go func(q string) {
		defer wg.Done()
		albums, err := spClient.SearchAlbums(q)
		if err != nil {
			log.Println("ERROR (sp SearchAlbums):", err.Error())
		}
		for i, a := range albums.Albums.Items {
			a.Name = truncate(a.Name, "...", 60)
			albums.Albums.Items[i] = a
		}
		s = albums
	}(q)
	wg.Wait()
	return i, g, s
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("index.html").ParseFiles("templates/index.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Render the template
	err = t.ExecuteTemplate(w, "base", map[string]interface{}{"Page": "home"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request, query string) {
	iKey, grKey, grSecret, err := parseYAML()
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client := &http.Client{}
	grClient := goodreads.Goodreads{Client: *client, Key: grKey, Secret: grSecret}
	spClient := spotify.Spotify{Client: client}
	imdbClient := imdb.Imdb{Client: client, Key: iKey}
	i, g, s := Search(query, imdbClient, grClient, spClient)
	// Since spotify: URIs are not trusted, have to pass a
	// URL function to the template to use in hrefs
	funcMap := template.FuncMap{
		"URL": func(q string) template.URL { return template.URL(query) },
		"spotifyImage": func(album spotify.Album) string {
			if len(album.Images) > 0 {
				return album.Images[len(album.Images)-1].URL
			}
			return ""
		},
	}
	t, err := template.New("search.html").Funcs(funcMap).ParseFiles("templates/search.html", "templates/base.html")
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Render the template
	err = t.ExecuteTemplate(w, "base", map[string]interface{}{"Movies": i, "Books": g, "Albums": s.Albums})
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SaveHandler(w http.ResponseWriter, r *http.Request) {
	t := r.FormValue("title")
	l := r.FormValue("link")
	m := r.FormValue("media_type")
	url := r.FormValue("image_url")
	err := insertEntry(t, l, m, url)
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/list", http.StatusFound)
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	e, err := readEntries()
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, fmt.Sprintf("Error reading entries: %v", err), http.StatusInternalServerError)
		return
	}
	m := buildEntryMap(e)
	// Create and parse Template
	t, err := template.New("list.html").ParseFiles("templates/list.html", "templates/base.html")
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Render the template
	t.ExecuteTemplate(w, "base", map[string]interface{}{"Entries": m, "Page": "list"})
}

func RemoveHandler(w http.ResponseWriter, r *http.Request) {
	i := r.FormValue("id")
	err := removeEntry(i)
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, fmt.Sprintf("Error reading entries: %v", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/list", http.StatusFound)
}

var validSearchPath = regexp.MustCompile("^/search/(.*)$")

func makeSearchHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validSearchPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[1])
	}
}
