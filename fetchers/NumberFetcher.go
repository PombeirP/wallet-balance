package fetchers

import (
	"errors"
	"io/ioutil"
	"strconv"
)

// NumberFetcher defines an interface for fetching numeric body responses from web APIs
type NumberFetcher interface {
	Fetch(url string) (result float64, err error)
}

// WebNumberFetcher implements the NumberFetcher interface for an HTTPClient
type WebNumberFetcher struct {
	client HTTPClient
}

// NewWebNumberFetcher returns an initialized instance of WebNumberFetcher
func NewWebNumberFetcher(client HTTPClient) *WebNumberFetcher {
	return &WebNumberFetcher{client}
}

// Fetch calls a web API and decodes the JSON response
func (fetcher *WebNumberFetcher) Fetch(url string) (result float64, err error) {
	resp, err := fetcher.client.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	bodyString := string(body)
	if resp.StatusCode >= 300 {
		if len(bodyString) > 0 {
			err = errors.New(bodyString)
		} else {
			err = errors.New(resp.Status)
		}

		return
	}

	result, err = strconv.ParseFloat(bodyString, 64)

	return
}
