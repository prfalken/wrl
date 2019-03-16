package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/kylelemons/go-gypsy/yaml"
)

func parseYAML() (imdbKey, grKey, grSecret string, err error) {
	config, err := yaml.ReadFile(*configFile)
	if err != nil {
		return
	}
	imdbKey, err = config.Get("imdb.key")
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

	return imdbKey, grKey, grSecret, nil
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
		if e.ID == id {
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
