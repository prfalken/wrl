package main

import (
	"flag"
	"github.com/kylelemons/go-gypsy/yaml"
	//"github.com/shawnps/gr"
	//"github.com/shawnps/rt"
	//"github.com/shawnps/sp"
	"log"
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
