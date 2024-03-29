package main

import (
	"flag"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

var (
	port        = flag.String("p", "8000", "Port number (default 8000)")
	configFile  = flag.String("c", "config.yml", "Config file (default config.yml)")
	entriesPath = flag.String("f", "entries.json", "Path to JSON storage file (default entries.json)")
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
}

func main() {
	flag.Parse()
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/search/", makeSearchHandler(SearchHandler))
	http.HandleFunc("/save", SaveHandler)
	http.HandleFunc("/list", ListHandler)
	http.HandleFunc("/remove", RemoveHandler)
	log.Println("Running on localhost:" + *port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+*port, nil))
}
