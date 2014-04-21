package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/gorilla/mux"
	"github.com/kylelemons/go-gypsy/yaml"
	_ "github.com/lib/pq"
	"github.com/shawnps/gr"
	"github.com/shawnps/rt"
	"github.com/shawnps/sp"
)

var (
	port       = flag.String("p", "8000", "Port number (default 8000)")
	configFile = flag.String("c", "config.yml", "Config file (default config.yml)")
)

type entry struct {
	Id       string
	Title    string
	Link     string
	ImageURL url.URL
}

type Entries struct {
	Movies []entry
	Books  []entry
	Albums []entry
}

func parseYAML() (rtKey, grKey, grSecret string, err error) {
	config, err := yaml.ReadFile(*configFile)
	if err != nil {
		return
	}
	rtKey, err = config.Get("rt")
	if err != nil {
		return
	}
	grKey, err = config.Get("gr.key")
	if err != nil {
		return
	}
	grSecret, err = config.Get("gr.secret")
	if err != nil {
		return
	}

	return rtKey, grKey, grSecret, nil
}

func writeJSON(e Entries, file string) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, b, 0755)
	if err != nil {
		return err
	}

	return nil
}

func readEntries() (Entries, error) {
	var e Entries
	b, err := ioutil.ReadFile("entries.json")
	if err != nil {
		return e, err
	}
	err = json.Unmarshal(b, &e)
	if err != nil {
		return e, err
	}

	return e, nil
}

func uuid() (string, error) {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return "", err
	}
	b := make([]byte, 16)
	f.Read(b)
	f.Close()
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return uuid, nil
}

func insertEntry(title, link, mediaType, imageURL string) error {
	if _, err := os.Stat("entries.json"); os.IsNotExist(err) {
		_, err := os.Create("entries.json")
		if err != nil {
			return err
		}
		err = writeJSON(Entries{}, "entries.json")
		if err != nil {
			return err
		}
	}
	e, err := readEntries()
	if err != nil {
		return err
	}
	url, err := url.Parse(imageURL)
	if err != nil {
		return err
	}
	id, err := uuid()
	if err != nil {
		return err
	}
	entry := entry{id, title, link, *url}
	switch mediaType {
	case "movie":
		e.Movies = append(e.Movies, entry)
	case "book":
		e.Books = append(e.Books, entry)
	case "album":
		e.Albums = append(e.Albums, entry)
	}
	err = writeJSON(e, "entries.json")
	if err != nil {
		return err
	}
	return nil
}

func truncate(s, suf string, l int) string {
	if len(s) < l {
		return s
	} else {
		return s[:l] + suf
	}
}

// Search Rotten Tomatoes, Goodreads, and Spotify.
func Search(q string, rtClient rt.RottenTomatoes, grClient gr.Goodreads, spClient sp.Spotify) (m []rt.Movie, g gr.GoodreadsResponse, s sp.SearchAlbumsResponse) {
	var wg sync.WaitGroup
	wg.Add(3)
	go func(q string) {
		defer wg.Done()
		movies, err := rtClient.SearchMovies(q)
		if err != nil {
			fmt.Println("ERROR (rt): ", err.Error())
		}
		for _, mov := range movies {
			mov.Title = truncate(mov.Title, "...", 60)
			m = append(m, mov)
		}
	}(q)
	go func(q string) {
		defer wg.Done()
		books, err := grClient.SearchBooks(q)
		if err != nil {
			fmt.Println("ERROR (gr): ", err.Error())
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
			fmt.Println("ERROR (sp): ", err.Error())
		}
		for i, a := range albums.Albums {
			a.Name = truncate(a.Name, "...", 60)
			albums.Albums[i] = a
		}
		s = albums
	}(q)
	wg.Wait()
	return m, g, s
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("index.html").ParseFiles("templates/index.html", "templates/base.html")
	if err != nil {
		log.Panic(err)
	}
	// Render the template
	err = t.ExecuteTemplate(w, "base", map[string]interface{}{"Page": "home"})
	if err != nil {
		log.Panic(err)
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	q := vars["query"]
	q, err := url.QueryUnescape(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	rtKey, grKey, grSecret, err := parseYAML()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	rtClient := rt.RottenTomatoes{rtKey}
	grClient := gr.Goodreads{grKey, grSecret}
	spClient := sp.Spotify{}
	m, g, s := Search(q, rtClient, grClient, spClient)
	// Since spotify: URIs are not trusted, have to pass a
	// URL function to the template to use in hrefs
	funcMap := template.FuncMap{
		"URL": func(q string) template.URL { return template.URL(q) },
	}
	t, err := template.New("search.html").Funcs(funcMap).ParseFiles("templates/search.html", "templates/base.html")
	if err != nil {
		log.Panic(err)
	}
	// Render the template
	err = t.ExecuteTemplate(w, "base", map[string]interface{}{"Movies": m, "Books": g, "Albums": s.Albums})
	if err != nil {
		log.Panic(err)
	}
}

func SaveHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", "user=postgres dbname=wrl sslmode=require")
	if err != nil {
		log.Println("Error opening db connection: " + err.Error())
		return
	}
	defer db.Close()
	t := r.FormValue("title")
	l := r.FormValue("link")
	m := r.FormValue("media_type")
	url := r.FormValue("image_url")
	err = insertEntry(t, l, m, url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/list", http.StatusFound)
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := readEntries()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading entries: %v", err), 500)
		return
	}
	// Create and parse Template
	t, err := template.New("list.html").ParseFiles("templates/list.html", "templates/base.html")
	if err != nil {
		log.Panic(err)
	}
	// Render the template
	t.ExecuteTemplate(w, "base", map[string]interface{}{"Entries": entries, "Page": "list"})
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/search/{query}", SearchHandler)
	r.HandleFunc("/save", SaveHandler)
	r.HandleFunc("/list", ListHandler)
	http.Handle("/", r)
	fmt.Println("Running on localhost:" + *port)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
