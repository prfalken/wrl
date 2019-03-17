package spotify

import (
	"os"
	"testing"

	"github.com/kylelemons/go-gypsy/yaml"
)

func TestSpotify(t *testing.T) {
	spClientID := ""
	spClientSecret := ""
	if os.Getenv("SPOTIFY_KEY") != "" && os.Getenv("SPOTIFY_SECRET") != "" {
		spClientID = os.Getenv("IMDB_CLIENTID")
		spClientSecret = os.Getenv("SPOTIFY_CLIENTSECRET")

	} else {
		config, err := yaml.ReadFile("config_test.yml")
		if err != nil {
			t.Log("Could not read config file")

		}
		spClientID, err = config.Get("spotify.clientID")
		if err != nil {
			t.Fatal("could not get spotify clientID from config file")
		}
		spClientSecret, err = config.Get("spotify.clientSecret")
		if err != nil {
			t.Fatal("could not get spotify clientSecret from config file")
		}
	}
	s := Spotify{ClientID: spClientID, ClientSecret: spClientSecret}
	_, err := s.SearchAlbums("Selling England by the Pound")

	if err != nil {
		t.Log(err)
		t.Fatalf("error with Spotify API call")
	}
}
