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
func Search(q string, rt rt.RottenTomatoes, gr gr.Goodreads, sp sp.Spotify) ([]rt.Movie, gr.GoodreadsResponse, sp.SearchAlbumsResponse) {
	m, err := rt.SearchMovies(q)
	if err != nil {
		fmt.Println("ERROR (rt): ", err.Error())
	}
	g, err := gr.SearchBooks(q)
	if err != nil {
		fmt.Println("ERROR (gr): ", err.Error())
	}
	s, err := sp.SearchAlbums(q)
	if err != nil {
		fmt.Println("ERROR (sp): ", err.Error())
	}
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
	rt := rt.RottenTomatoes{rtKey}
	gr := gr.Goodreads{grKey, grSecret}
	sp := sp.Spotify{}
	m, g, s := Search(q, rt, gr, sp)
	t, err := template.New("search.html").ParseFiles("templates/search.html")
	if err != nil {
		log.Panic(err)
	}
	// Render the template
	err = t.Execute(w, map[string]interface{}{"Movies": m, "Books": g, "Albums": s})
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
