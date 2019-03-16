package goodreads

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
)

const apiURL = "https://www.goodreads.com/"

type Goodreads struct {
	Client http.Client
	Key    string
	Secret string
}

type Author struct {
	ID   int    `xml:"id"`
	Name string `xml:"name"`
}

type GoodreadsResponse struct {
	Search Search `xml:"search"`
	Query  string `xml:"query"`
}

type Search struct {
	ResultsStart     int     `xml:"results-start"`
	ResultsEnd       int     `xml:"results-end"`
	TotalResults     int     `xml:"total-results"`
	Source           string  `xml:"source"`
	QueryTimeSeconds float64 `xml:"query-time-seconds"`
	Works            []Work  `xml:"results>work"`
}

type Results struct {
	Works []Work `xml:"work"`
}

type Work struct {
	BooksCount int `xml:"books_count"`
	ID         int `xml:"id"`
	/* TODO: figure out how to unmarshal XML into interface{}
	OriginalPublicationDay   interface{}      `xml:"original_publication_day,omitempty"`
	OriginalPublicationMonth interface{}      `xml:"original_publication_month,omitempty"`
	OriginalPublicationYear  interface{}      `xml:"original_publication_year,omitempty"`
	*/
	RatingsCount     int      `xml:"ratings_count"`
	TextReviewsCount int      `xml:"text_reviews_count"`
	AverageRating    float64  `xml:"average_rating"`
	BestBook         BestBook `xml:"best_book"`
}

type BestBook struct {
	ID            int    `xml:"id"`
	Title         string `xml:"title"`
	Author        Author `xml:"author"`
	ImageURL      string `xml:"image_url"`
	SmallImageURL string `xml:"small_image_url"`
}

func (g *Goodreads) getRequest(params map[string]string, endpoint string) ([]byte, error) {
	v := url.Values{}
	v.Set("key", g.Key)
	for key, val := range params {
		v.Set(key, val)
	}
	u := apiURL + endpoint + "?" + v.Encode()
	resp, err := g.Client.Get(u)
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

func (g *Goodreads) SearchBooks(q string) (GoodreadsResponse, error) {
	p := map[string]string{"q": q}
	var gr GoodreadsResponse
	resp, err := g.getRequest(p, "search.xml")
	if err != nil {
		return GoodreadsResponse{}, err
	}
	err = xml.Unmarshal(resp, &gr)
	if err != nil {
		return GoodreadsResponse{}, err
	}
	return gr, nil
}
