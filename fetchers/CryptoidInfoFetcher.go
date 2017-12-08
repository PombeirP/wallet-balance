package fetchers

import (
	"fmt"
)

// CryptoidInfoFetcher fetches the balance and exchange rate of several altcoins on chainz.cryptoid.info
type CryptoidInfoFetcher struct {
	currency   string
	apiFetcher NumberFetcher
}

// NewCryptoidInfoFetcher creates an instance of CryptoidInfoFetcher for a specified altcoin from an HTTP client instance
func NewCryptoidInfoFetcher(currency string, client HTTPClient) *CryptoidInfoFetcher {
	numberFetcher := NewWebNumberFetcher(client)
	return &CryptoidInfoFetcher{currency, numberFetcher}
}

// FetchBalance retrieves the aggregate balances on chainz.cryptoid.info for the provided addresses
func (fetcher *CryptoidInfoFetcher) FetchBalance(addresses []string, apiKey string, balance *float64, err *error, done chan<- bool) {
	*err = nil
	*balance = 0.

	balances := make(chan float64)
	errorsChan := make(chan error)

	for _, address := range addresses {
		url := fmt.Sprintf("https://chainz.cryptoid.info/%s/api.dws?q=getbalance&key=%s&a=%s", fetcher.currency, apiKey, address)
		go fetcher.apiFetcher.Fetch(url, balances, errorsChan)
	}

	for _ = range addresses {
		select {
		case *err = <-errorsChan:
		case partialBalance := <-balances:
			*balance += partialBalance
		}
	}

	done <- true
}

// FetchExchangeRate retrieves the exchange rate for BTC in `targetCurrency`
func (fetcher *CryptoidInfoFetcher) FetchExchangeRate(apiKey string, targetCurrency string, exchangeRate *float64, err *error, done chan<- bool) {
	*exchangeRate = 0.

	exchangeRates := make(chan float64)
	errorsChan := make(chan error)

	url := fmt.Sprintf("https://chainz.cryptoid.info/%s/api.dws?q=ticker.%s&key=%s", fetcher.currency, targetCurrency, apiKey)
	go fetcher.apiFetcher.Fetch(url, exchangeRates, errorsChan)

	select {
	case *err = <-errorsChan:
	case *exchangeRate = <-exchangeRates:
	}

	done <- true
}
