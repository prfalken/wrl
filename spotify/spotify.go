package spotify

// TODO: use https://github.com/zmb3/spotify

import (
	"context"
	"net/http"

	spotifyapp "github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
)

type Spotify struct {
	Client       *http.Client
	ClientID     string
	ClientSecret string
}

func (s *Spotify) SearchAlbums(q string) (albums []spotifyapp.SimpleAlbum, err error) {

	config := &clientcredentials.Config{
		ClientID:     s.ClientID,
		ClientSecret: s.ClientSecret,
		TokenURL:     spotifyapp.TokenURL,
	}
	token, err := config.Token(context.Background())
	if err != nil {
		return nil, err
	}

	client := spotifyapp.Authenticator{}.NewClient(token)
	// search for playlists and albums containing "holiday"
	results, err := client.Search(q, spotifyapp.SearchTypeAlbum)
	if err != nil {
		return nil, err
	}

	// handle album results
	return results.Albums.Albums, nil
}
