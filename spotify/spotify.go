package spotify

// TODO: use https://github.com/zmb3/spotify

import (
	"net/http"

	spotifyapp "github.com/zmb3/spotify"
)

type Spotify struct {
	Client       *http.Client
	ClientID     string
	ClientSecret string
}

type Album struct {
	Name         string
	Released     string  `json:"released,omitempty"`
	Length       float64 `json:"length,omitempty"`
	Href         string
	Availability struct {
		Territories string
	}
}

func (s *Spotify) SearchAlbums(q string) (sr spotifyapp.SearchResult, err error) {

	// redirectURL := "https://wrl.falken.dev"
	// auth := spotifyapp.NewAuthenticator(redirectURL, spotifyapp.ScopeUserReadPrivate)
	// auth.SetAuthInfo(s.ClientID, s.ClientSecret)
	// url := auth.AuthURL("xxx")

	// result, err := spotifyapp.Search(query, spotifyapp.SearchTypeAlbum)
	// if err != nil {
	// 	return sr, err
	// }
	return
}

// func redirectHandler(w http.ResponseWriter, r *http.Request) {
//       // use the same state string here that you used to generate the URL
//       token, err := auth.Token(state, r)
//       if err != nil {
//             http.Error(w, "Couldn't get token", http.StatusNotFound)
//             return
//       }
//       // create a client using the specified token
//       client := auth.NewClient(token)

//       // the client can now be used to make authenticated requests
// }
