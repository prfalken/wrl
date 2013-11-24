package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/shawnps/gr"
	"github.com/shawnps/rt"
	"github.com/shawnps/sp"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var (
	port       = flag.String("p", "8000", "Port number (default 8000)")
	configFile = flag.String("c", "config.yml", "Config file (default config.yml)")
)

func getYAMLString(n yaml.Node, key string) string {
	return strings.TrimSpace(n.(yaml.Map)[key].(yaml.Scalar).String())
}

func parseYAML() (rtKey, grKey, grSecret string) {
	config, err := yaml.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	configRoot, _ := config.Root.(yaml.Map)
	rtKey = configRoot["rt"].(yaml.Scalar).String()
	g := configRoot["gr"]
	grKey = getYAMLString(g, "key")
	grSecret = getYAMLString(g, "secret")

	return rtKey, grKey, grSecret
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
			m = append(m, mov)
		}
	}(q)
	go func(q string) {
		defer wg.Done()
		books, err := grClient.SearchBooks(q)
		if err != nil {
			fmt.Println("ERROR (gr): ", err.Error())
		}
		g = books
	}(q)
	go func(q string) {
		defer wg.Done()
		albums, err := spClient.SearchAlbums(q)
		if err != nil {
			fmt.Println("ERROR (sp): ", err.Error())
		}
		s = albums
	}(q)
	wg.Wait()
	return m, g, s
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	q := vars["query"]
	q, err := url.QueryUnescape(q)
	if err != nil {
		log.Panic(err)
	}
	rtKey, grKey, grSecret := parseYAML()
	rtClient := rt.RottenTomatoes{rtKey}
	grClient := gr.Goodreads{grKey, grSecret}
	spClient := sp.Spotify{}
	m, g, s := Search(q, rtClient, grClient, spClient)
	// Since spotify: URIs are not trusted, have to pass a
	// URL function to the template to use in hrefs
	funcMap := template.FuncMap{
		"URL": func(q string) template.URL { return template.URL(q) },
	}
	t, err := template.New("search.html").Funcs(funcMap).ParseFiles("templates/search.html")
	if err != nil {
		log.Panic(err)
	}
	// Render the template
	err = t.Execute(w, map[string]interface{}{"Movies": m, "Books": g, "Albums": s.Albums})
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/search/{query}", SearchHandler)
	http.Handle("/", r)
	fmt.Println("Running on localhost:" + *port)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
