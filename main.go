package main

import (
	"flag"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/shawnps/gr"
	"github.com/shawnps/rt"
	"github.com/shawnps/sp"
	"log"
	"net/http"
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
func Search(q string) {
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	rtKey, grKey, grSecret := parseYAML()
	rt := rt.RottenTomatoes{rtKey}
	gr := gr.Goodreads{grKey, grSecret}
	sp := sp.Spotify{}
	fmt.Println(rt)
	fmt.Println(gr)
	fmt.Println(sp)
	fmt.Fprintf(w, "Hello")
}

func main() {
	http.HandleFunc("/", HomeHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
