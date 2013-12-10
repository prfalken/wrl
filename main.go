package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kylelemons/go-gypsy/yaml"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shawnps/gr"
	"github.com/shawnps/rt"
	"github.com/shawnps/sp"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
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

func createDb() {
	db, err := sql.Open("sqlite3", "./watchreadlisten.db")
	if err != nil {
		log.Println("Error opening or creating deploy_log.db: " + err.Error())
		return
	}
	defer db.Close()
	sql := `create table if not exists entries (id integer not null primary key autoincrement, title text, link text, media_type text, timestamp datetime default current_timestamp);`
	_, err = db.Exec(sql)
	if err != nil {
		log.Println("Error creating logs table: " + err.Error())
		return
	}
}

func insertEntry(db sql.DB, title, link, mediaType string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error connecting to database: " + err.Error())
		return err
	}
	stmt, err := tx.Prepare("insert into entries(title, link, media_type) values(?, ?, ?)")
	if err != nil {
		log.Println("Error inserting to database: " + err.Error())
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(title, link, mediaType)
	if err != nil {
		log.Println("Error executing statement on database: " + err.Error())
		return err
	}
	tx.Commit()
	return nil
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
	t, err := template.New("index.html").ParseFiles("templates/index.html", "templates/base.html")
	if err != nil {
		log.Panic(err)
	}
	// Render the template
	err = t.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Panic(err)
	}
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
	db, err := sql.Open("sqlite3", "./watchreadlisten.db")
	if err != nil {
		log.Println("Error opening sqlite db: " + err.Error())
		return
	}
	defer db.Close()
	t := r.FormValue("title")
	l := r.FormValue("link")
	m := r.FormValue("media_type")
	err = insertEntry(*db, t, l, m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./watchreadlisten.db")
	if err != nil {
		log.Println("Error opening sqlite db: " + err.Error())
		return
	}
	defer db.Close()
	q := "select id, title, link, media_type, timestamp from entries"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("Error querying sqlite db: " + err.Error())
		return
	}
	type Entry struct {
		Id        int
		Title     string
		Link      string
		MediaType string `db:"media_type"`
		Timestamp time.Time
	}
	entries := map[string][]Entry{}
	defer rows.Close()
	for rows.Next() {
		e := Entry{}
		var id int
		var title, link, mediaType string
		var timestamp time.Time
		rows.Scan(&id, &title, &link, &mediaType, &timestamp)
		e.Id = id
		e.Title = title
		e.Link = link
		e.MediaType = mediaType
		e.Timestamp = timestamp
		entries[mediaType] = append(entries[mediaType], e)
	}
	// Create and parse Template
	t, err := template.New("list.html").ParseFiles("templates/list.html", "templates/base.html")
	if err != nil {
		log.Panic(err)
	}
	// Render the template
	t.ExecuteTemplate(w, "base", map[string]interface{}{"Entries": entries})
}

func main() {
	createDb()
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/search/{query}", SearchHandler)
	r.HandleFunc("/save", SaveHandler)
	r.HandleFunc("/list", ListHandler)
	http.Handle("/", r)
	fmt.Println("Running on localhost:" + *port)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
