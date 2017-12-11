package fetchers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

// JSONFetcher defines an interface for fetching JSON responses from web APIs
type JSONFetcher interface {
	Fetch(url string, response interface{}) error
}

// EtherscanJSONFetcher implements the JSONFetcher interface for an HTTPClient to parse an etherscan.io response
type EtherscanJSONFetcher struct {
	client HTTPClient
}

// NewEtherscanJSONFetcher returns an initialized instance of EtherscanJSONFetcher
func NewEtherscanJSONFetcher(client HTTPClient) *EtherscanJSONFetcher {
	return &EtherscanJSONFetcher{client}
}

// Fetch calls a web API and decodes the JSON response
func (fetcher *EtherscanJSONFetcher) Fetch(url string, response interface{}) (err error) {
	resp, err := fetcher.client.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var untypedResponse map[string]interface{}
	err = json.Unmarshal([]byte(body), &untypedResponse)
	if err != nil {
		return
	}

	if untypedResponse["status"].(string) != "1" {
		err = errors.New(untypedResponse["message"].(string))
		return
	}

	json.Unmarshal([]byte(body), response)

	return
}
