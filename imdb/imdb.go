package imdb

import (
	"net/http"

	omdb "github.com/kenshaw/imdb"
)

type Imdb struct {
	Client *http.Client
	Key    string
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
