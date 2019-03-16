package sp

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	apiURL = "https://api.spotify.com"
)

type Spotify struct {
	Client *http.Client
}

// FloatString takes a string and converts it to a float64.
// The API returns strings for some things that one would expect
// to be numbers (for example, popularity))
type FloatString string

func (f *FloatString) UnmarshalJSON(i interface{}) float64 {
	s := i.(string)
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

// IntString is the same as FloatString, but for integers.
type IntString string

func (f *IntString) UnmarshalJSON(i interface{}) int64 {
	s := i.(string)
	n, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

type Image struct {
	Height int
	URL    string
	Width  int
}

type ExternalID struct {
	Type string
	ID   IntString
}

type Artist struct {
	Href       string
	Name       string
	Popularity FloatString `json:"popularity,omitempty"`
}

type Album struct {
	Name         string
	Released     string       `json:"released,omitempty"`
	Popularity   FloatString  `json:"popularity,omitempty"`
	ExternalIDs  []ExternalID `json:"external-ids"`
	Length       float64      `json:"length,omitempty"`
	Href         string
	Artists      []Artist `json:"artists,omitempty"`
	Availability struct {
		Territories string
	}
	Images []Image
}

type Info struct {
	NumResults int `json:"num_results"`
	Limit      int
	Offset     int
	Query      string
	Type       string
	Page       int
}

type SearchAlbumsResponse struct {
	Info   Info
	Albums struct {
		Items []Album
	}
}

type SearchArtistsResponse struct {
	Info    Info
	Artists []Artist
}

type Track struct {
	Album        Album
	Name         string
	ExternalIDs  []ExternalID `json:"external-ids"`
	Popularity   FloatString
	Explicit     bool `json:"explicit,omitempty"`
	Length       float64
	Href         string
	Artists      []Artist
	TrackNumber  IntString `json:"track-number"`
	Availability struct {
		Territories string
	} `json:"availability,omitempty"`
}

type SearchTracksResponse struct {
	Info   Info
	Tracks []Track
}

type LookupArtistResponse struct {
	Info struct {
		Type string
	}
	Artist Artist
}

type LookupTrackResponse struct {
	Info struct {
		Type string
	}
	Track Track
}

type LookupAlbumResponse struct {
	Info struct {
		Type string
	}
	Album struct {
		ArtistID    string `json:"artist-id"`
		Name        string
		Artist      string
		ExternalIDs []struct {
			Type string
			ID   string
		} `json:"external-ids"`
		Released string
		Tracks   []struct {
			Available bool
			Href      string
			Artists   []struct {
				Href string
				Name string
			}
			Name string
		}
	}
}

func (r *Spotify) getRequest(params map[string]string, endpoint string) ([]byte, error) {
	v := url.Values{}
	for key, val := range params {
		v.Set(key, val)
	}
	u := apiURL + endpoint + "?" + v.Encode()
	resp, err := r.Client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *Spotify) SearchAlbums(q string) (SearchAlbumsResponse, error) {
	p := map[string]string{"q": q, "type": "album"}
	e := "/v1/search"
	resp, err := r.getRequest(p, e)
	var s SearchAlbumsResponse
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(resp, &s)
	if err != nil {
		return s, err
	}
	return s, nil
}

func (r *Spotify) SearchArtists(q string) (SearchArtistsResponse, error) {
	p := map[string]string{"q": q}
	e := "/search/1/artist.json"
	resp, err := r.getRequest(p, e)
	var s SearchArtistsResponse
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(resp, &s)
	if err != nil {
		return s, err
	}
	return s, nil
}

func (r *Spotify) SearchTracks(q string) (SearchTracksResponse, error) {
	p := map[string]string{"q": q}
	e := "/search/1/track.json"
	resp, err := r.getRequest(p, e)
	var s SearchTracksResponse
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(resp, &s)
	if err != nil {
		return s, err
	}
	return s, nil
}

func (r *Spotify) LookupArtist(uri string) (LookupArtistResponse, error) {
	p := map[string]string{"uri": uri}
	e := "/lookup/1/.json"
	resp, err := r.getRequest(p, e)
	var l LookupArtistResponse
	if err != nil {
		return l, err
	}
	err = json.Unmarshal(resp, &l)
	if err != nil {
		return l, err
	}
	return l, nil
}

func (r *Spotify) LookupAlbum(uri string) (LookupAlbumResponse, error) {
	p := map[string]string{"uri": uri}
	e := "/lookup/1/.json"
	resp, err := r.getRequest(p, e)
	var l LookupAlbumResponse
	if err != nil {
		return l, err
	}
	err = json.Unmarshal(resp, &l)
	if err != nil {
		return l, err
	}
	return l, nil
}

func (r *Spotify) LookupTrack(uri string) (LookupTrackResponse, error) {
	p := map[string]string{"uri": uri}
	e := "/lookup/1/.json"
	resp, err := r.getRequest(p, e)
	var l LookupTrackResponse
	if err != nil {
		return l, err
	}
	err = json.Unmarshal(resp, &l)
	if err != nil {
		return l, err
	}
	return l, nil
}
