package imdb

import (
	"net/http"

	omdb "github.com/kenshaw/imdb"
)

type Imdb struct {
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

func (i *Imdb) SearchMovies(q string) (movies []*omdb.MovieResult, err error) {
	cl := omdb.New(i.Key)
	response, err := cl.Search(q, "")
	if err != nil {
		return nil, err
	}

	for _, movie := range response.Search {
		movieResponse, err := cl.MovieByImdbID(movie.ImdbID)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movieResponse)
	}

	return movies, nil
}
