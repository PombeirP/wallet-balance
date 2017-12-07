package fetchers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// JSONFetcher defines an interface for fetching JSON responses from web APIs
type JSONFetcher interface {
	Fetch(url string, response interface{}, responseReadyChan chan<- bool, errorsChan chan<- error)
}

// EtherscanJSONFetcher implements the JSONFetcher interface for an http.Client to parse an etherscan.io response
type EtherscanJSONFetcher struct {
	client *http.Client
}

// NewEtherscanJSONFetcher returns an initialized instance of EtherscanJSONFetcher
func NewEtherscanJSONFetcher(client *http.Client) *EtherscanJSONFetcher {
	return &EtherscanJSONFetcher{client}
}

// Fetch calls a web API and decodes the JSON response
func (fetcher *EtherscanJSONFetcher) Fetch(url string, response interface{}, responseReadyChan chan<- bool, errorsChan chan<- error) {
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

	json.Unmarshal([]byte(body), response)

	var untypedResponse map[string]interface{}
	err = json.Unmarshal([]byte(body), &untypedResponse)
	if err != nil {
		errorsChan <- err
		return
	}

	if untypedResponse["status"].(string) != "1" {
		errorsChan <- errors.New(untypedResponse["message"].(string))
		return
	}

	responseReadyChan <- true
}
