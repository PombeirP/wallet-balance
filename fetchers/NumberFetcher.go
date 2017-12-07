package fetchers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

// NumberFetcher defines an interface for fetching numeric body responses from web APIs
type NumberFetcher interface {
	Fetch(url string, resultChan chan<- float64, errorsChan chan<- error)
}

// WebNumberFetcher implements the NumberFetcher interface for an http.Client
type WebNumberFetcher struct {
	client *http.Client
}

// NewWebNumberFetcher returns an initialized instance of WebNumberFetcher
func NewWebNumberFetcher(client *http.Client) *WebNumberFetcher {
	return &WebNumberFetcher{client}
}

// Fetch calls a web API and decodes the JSON response
func (fetcher *WebNumberFetcher) Fetch(url string, resultChan chan<- float64, errorsChan chan<- error) {
	resp, err := fetcher.client.Get(url)
	if err != nil {
		errorsChan <- err
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorsChan <- err
		return
	}

	bodyString := string(body)
	if resp.StatusCode >= 300 {
		if len(bodyString) > 0 {
			errorsChan <- errors.New(bodyString)
		} else {
			errorsChan <- errors.New(resp.Status)
		}

		return
	}

	value, err := strconv.ParseFloat(bodyString, 64)
	if err != nil {
		errorsChan <- err
		return
	}

	resultChan <- value
}
