package fetchers

import (
	"fmt"
	"strings"
	"sync"
)

const (
	satoshi = 100000000. // 10^8
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
func (fetcher *BlockchainInfoFetcher) FetchBalance(addresses []string, apiKey string, balance *float64, err *error, done *sync.WaitGroup) {
	url := fmt.Sprintf("https://blockchain.info/q/addressbalance/%s", strings.Join(addresses, "%7C" /*|*/))

	if *balance, *err = fetcher.apiFetcher.Fetch(url); *err == nil {
		*balance = *balance / satoshi
	}

	done.Done()
}

// FetchExchangeRate retrieves the exchange rate for BTC in `targetCurrency`
func (fetcher *BlockchainInfoFetcher) FetchExchangeRate(apiKey string, targetCurrency string, exchangeRate *float64, err *error, done *sync.WaitGroup) {
	url := fmt.Sprintf("https://blockchain.info/tobtc?currency=%s&value=1", targetCurrency)

	if *exchangeRate, *err = fetcher.apiFetcher.Fetch(url); *err == nil {
		*exchangeRate = 1. / *exchangeRate
	}

	done.Done()
}
