package fetchers

import (
	"fmt"
	"sync"
)

// CryptoidInfoFetcher fetches the balance and exchange rate of several altcoins on https://chainz.cryptoid.info/
type CryptoidInfoFetcher struct {
	currency   string
	apiFetcher NumberFetcher
}

// NewCryptoidInfoFetcher creates an instance of CryptoidInfoFetcher for a specified altcoin from an HTTP client instance
func NewCryptoidInfoFetcher(currency string, client HTTPClient) *CryptoidInfoFetcher {
	numberFetcher := NewWebNumberFetcher(client)
	return &CryptoidInfoFetcher{currency, numberFetcher}
}

// FetchBalance retrieves the aggregate balances on https://chainz.cryptoid.info/ for the provided addresses
func (fetcher *CryptoidInfoFetcher) FetchBalance(addresses []string, apiKey string, balance *float64, err *error, done *sync.WaitGroup) {
	*err = nil
	*balance = 0.

	balancesChan := make(chan float64)
	errorsChan := make(chan error)

	for _, address := range addresses {
		url := fmt.Sprintf("https://chainz.cryptoid.info/%s/api.dws?q=getbalance&key=%s&a=%s", fetcher.currency, apiKey, address)
		go func() {
			balance, err := fetcher.apiFetcher.Fetch(url)
			if err == nil {
				balancesChan <- balance
			} else {
				errorsChan <- err
			}
		}()
	}

	for range addresses {
		select {
		case *err = <-errorsChan:
		case partialBalance := <-balancesChan:
			*balance += partialBalance
		}
	}

	done.Done()
}

// FetchExchangeRate retrieves the exchange rate for BTC in `targetCurrency`
func (fetcher *CryptoidInfoFetcher) FetchExchangeRate(apiKey string, targetCurrency string, exchangeRate *float64, err *error, done *sync.WaitGroup) {
	url := fmt.Sprintf("https://chainz.cryptoid.info/%s/api.dws?q=ticker.%s&key=%s", fetcher.currency, targetCurrency, apiKey)
	*exchangeRate, *err = fetcher.apiFetcher.Fetch(url)

	done.Done()
}
