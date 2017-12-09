package fetchers

import (
	"fmt"
	"strings"
)

// BlockchainInfoFetcher fetches the balance and exchange rate of BTC on https://blockchain.info/
type BlockchainInfoFetcher struct {
	apiFetcher NumberFetcher
}

// NewBlockchainInfoFetcher creates an instance of BlockchainInfoFetcher from an HTTP client instance
func NewBlockchainInfoFetcher(client HTTPClient) *BlockchainInfoFetcher {
	numberFetcher := NewWebNumberFetcher(client)
	return &BlockchainInfoFetcher{numberFetcher}
}

// FetchBalance retrieves the aggregate balances on https://blockchain.info/ for the provided addresses
func (fetcher *BlockchainInfoFetcher) FetchBalance(addresses []string, apiKey string, balance *float64, err *error, done chan<- bool) {
	balances := make(chan float64)
	errorsChan := make(chan error)

	url := fmt.Sprintf("https://blockchain.info/q/addressbalance/%s", strings.Join(addresses, "%7C" /*|*/))
	go fetcher.apiFetcher.Fetch(url, balances, errorsChan)

	select {
	case *err = <-errorsChan:
	case *balance = <-balances:
		*balance = *balance / 100000000.
	}

	done <- true
}

// FetchExchangeRate retrieves the exchange rate for BTC in `targetCurrency`
func (fetcher *BlockchainInfoFetcher) FetchExchangeRate(apiKey string, targetCurrency string, exchangeRate *float64, err *error, done chan<- bool) {
	*exchangeRate = 0.

	exchangeRates := make(chan float64)
	errorsChan := make(chan error)

	url := fmt.Sprintf("https://blockchain.info/tobtc?currency=%s&value=1", targetCurrency)
	go fetcher.apiFetcher.Fetch(url, exchangeRates, errorsChan)

	select {
	case *err = <-errorsChan:
	case invExchangeRate := <-exchangeRates:
		*exchangeRate = 1. / invExchangeRate
	}

	done <- true
}
