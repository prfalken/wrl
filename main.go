package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/shawnps/gr"
	"github.com/shawnps/rt"
	"github.com/shawnps/sp"
)

var (
	port        = flag.String("p", "8000", "Port number (default 8000)")
	configFile  = flag.String("c", "config.yml", "Config file (default config.yml)")
	entriesPath = flag.String("f", "entries.json", "Path to JSON storage file (default entries.json)")
)

type Entry struct {
	Id       string
	Title    string
	Link     string
	ImageURL url.URL
	Type     string
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

func writeJSON(e []Entry, file string) error {
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

func buildEntryMap(entries []Entry) map[string][]Entry {
	m := map[string][]Entry{}
	for _, e := range entries {
		k := strings.Title(e.Type)
		m[k] = append(m[k], e)
	}
	return m
}

func readEntries() ([]Entry, error) {
	var e []Entry
	b, err := ioutil.ReadFile(*entriesPath)
	if err != nil {
		return e, err
	}
	if len(b) == 0 {
		return []Entry{}, nil
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
	if _, err := os.Stat(*entriesPath); os.IsNotExist(err) {
		_, err := os.Create(*entriesPath)
		if err != nil {
			return err
		}
		err = writeJSON([]Entry{}, *entriesPath)
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
	entry := Entry{id, title, link, *url, mediaType}
	e = append(e, entry)
	err = writeJSON(e, *entriesPath)
	if err != nil {
		return err
	}
	return nil
}

func removeEntry(id string) error {
	entries, err := readEntries()
	if err != nil {
		return err
	}
	for i, e := range entries {
		if e.Id == id {
			entries = append(entries[:i], entries[i+1:]...)
		}
	}
	return writeJSON(entries, *entriesPath)
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
	rtKey, grKey, grSecret, err := parseYAML()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rtClient := rt.RottenTomatoes{rtKey}
	grClient := gr.Goodreads{grKey, grSecret}
	spClient := sp.Spotify{}
	m, g, s := Search(query, rtClient, grClient, spClient)
	// Since spotify: URIs are not trusted, have to pass a
	// URL function to the template to use in hrefs
	funcMap := template.FuncMap{
		"URL": func(q string) template.URL { return template.URL(query) },
	}
	t, err := template.New("search.html").Funcs(funcMap).ParseFiles("templates/search.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Render the template
	err = t.ExecuteTemplate(w, "base", map[string]interface{}{"Movies": m, "Books": g, "Albums": s.Albums})
	if err != nil {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/list", http.StatusFound)
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	e, err := readEntries()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading entries: %v", err), http.StatusInternalServerError)
		return
	}
	m := buildEntryMap(e)
	// Create and parse Template
	t, err := template.New("list.html").ParseFiles("templates/list.html", "templates/base.html")
	if err != nil {
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

func main() {
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/search/", makeSearchHandler(SearchHandler))
	http.HandleFunc("/save", SaveHandler)
	http.HandleFunc("/list", ListHandler)
	http.HandleFunc("/remove", RemoveHandler)
	fmt.Println("Running on localhost:" + *port)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
