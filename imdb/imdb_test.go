package imdb

import (
	"testing"

	"github.com/kylelemons/go-gypsy/yaml"
)

func TestImdb(t *testing.T) {
	config, err := yaml.ReadFile("config_test.yml")
	if err != nil {
		t.Fatal("Could not read config file")
	}
	imdbKey, err := config.Get("imdb.key")
	if err != nil {
		t.Fatal("could not get imdb key from config file")
	}

	i := Imdb{Key: imdbKey}
	resp, err := i.SearchMovies("Home Alone")

	if err != nil {
		t.Log(err)
		t.Fatalf("error with Omdb API call")
	}

	if len(resp) == 0 {
		t.Fatal("Imdb search for Home alone returned no results")
	}
}
