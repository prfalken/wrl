package rt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	apiURL = "http://api.rottentomatoes.com/api/public/v1.0"
)

type RottenTomatoes struct {
	Client *http.Client
	Key    string
}

type Ratings struct {
	CriticsRating  string `json:"critics_rating,omitempty"`
	CriticsScore   *int   `json:"critics_score,omitempty"`
	AudienceRating string `json:"audience_rating,omitempty"`
	AudienceScore  int    `json:"audience_score,omitempty"`
}

type Actor struct {
	Name       string
	Id         string
	Characters []string
}

type Movie struct {
	Id               interface{}
	Title            string
	Year             interface{}       `json:"year,omitempty"`
	MPAARating       string            `json:"mpaa_rating"`
	Runtime          interface{}       `json:"runtime,omitempty"`
	CriticsConsensus string            `json:"critics_consensus"`
	ReleaseDates     map[string]string `json:"release_dates"`
	Ratings          Ratings
	Synopsis         string
	Posters          map[string]string
	AbridgedCast     []Actor           `json:"abridged_cast"`
	AlternateIds     map[string]string `json:"alternate_ids"`
	Links            map[string]string
}

type MovieSearchResponse struct {
	Total        int `json:"total,omitempty"`
	Movies       []Movie
	Links        map[string]string
	LinkTemplate string
}

func (r *RottenTomatoes) getRequest(params map[string]string, endpoint string) ([]byte, error) {
	v := url.Values{}
	v.Set("apikey", r.Key)
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

func convertMovies(movies []Movie) ([]Movie, error) {
	for i, mov := range movies {
		movie := &movies[i]
		// IDs in the list movies response are strings,
		// so we convert them to ints
		id, err := strconv.Atoi(mov.Id.(string))
		if err != nil {
			return nil, err
		}
		movie.Id = id
		// The Runtime is returned as an empty string "" when
		// it is not available, so we have to discover its type and
		// set it to nil if it's a string.
		switch mov.Runtime.(type) {
		case int:
			movie.Runtime = mov.Runtime.(int)
		case string:
			movie.Runtime = nil
		}
		// Same for Year
		switch mov.Year.(type) {
		case int:
			movie.Year = mov.Year.(int)
		case string:
			movie.Year = nil
		}
		if *movie.Ratings.CriticsScore == -1 {
			movie.Ratings.CriticsScore = nil
		}
	}
	return movies, nil
}

func movieListRequest(body []byte) ([]Movie, error) {
	if bytes.Equal(body, []byte("<h1>Developer Inactive</h1>")) {
		return nil, fmt.Errorf("Developer Inactive: check RottenTomatoes API credentials")
	}
	var m MovieSearchResponse
	err := json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}
	movies, err := convertMovies(m.Movies)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *RottenTomatoes) BoxOffice(c string) ([]Movie, error) {
	p := map[string]string{"country": c}
	e := "/lists/movies/box_office.json"

	resp, err := r.getRequest(p, e)
	if err != nil {
		return nil, err
	}
	movies, err := movieListRequest(resp)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *RottenTomatoes) OpeningMovies(c string) ([]Movie, error) {
	p := map[string]string{"country": c}
	e := "/lists/movies/opening.json"

	resp, err := r.getRequest(p, e)
	if err != nil {
		return nil, err
	}
	movies, err := movieListRequest(resp)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *RottenTomatoes) SearchMovies(q string) ([]Movie, error) {
	p := map[string]string{"q": q}
	e := "/movies.json"
	resp, err := r.getRequest(p, e)
	if err != nil {
		return nil, err
	}
	movies, err := movieListRequest(resp)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *RottenTomatoes) GetMovie(id string) (Movie, error) {
	p := map[string]string{"id": id}
	e := fmt.Sprintf("/movies/%s.json", id)
	resp, err := r.getRequest(p, e)
	if err != nil {
		return Movie{}, err
	}
	var m Movie
	err = json.Unmarshal(resp, &m)
	if err != nil {
		return Movie{}, err
	}
	// Individual Movie responses contain numeric Ids and not strings
	// like the list movies response, so we have to convert it here
	m.Id = int(m.Id.(float64))
	return m, nil
}