package main

import (
	"fmt"

	"github.com/PombeirP/wallet-balance/fetchers"
)

// cryptoCurrencyMap maps cryptoCurrencyTickerSymbol values to internal symbols used by external APIs
var cryptoCurrencyMap map[cryptoCurrencyTickerSymbol]string

func init() {
	cryptoCurrencyMap = map[cryptoCurrencyTickerSymbol]string{
		btc:  "btc",
		eth:  "eth",
		ltc:  "ltc",
		dash: "dash",
		uno:  "uno",
		bcc:  "bcc",
	}
}

// CryptoCurrencyInfoFetcherCreator defines the interface for a factory that creates a fetchers.CryptoCurrencyInfoFetcher based on a currency symbol
type CryptoCurrencyInfoFetcherCreator interface {
	Create(symbol cryptoCurrencyTickerSymbol) (fetchers.CryptoCurrencyInfoFetcher, error)
}

// CryptoCurrencyInfoHTTPFetcherCreator implements a factory that creates a fetchers.CryptoCurrencyInfoFetcher based on a currency symbol and an HTTP client
type CryptoCurrencyInfoHTTPFetcherCreator struct {
	client fetchers.HTTPClient
}

// NewCryptoCurrencyInfoHTTPFetcherCreator creates a CryptoCurrencyInfoHTTPFetcherCreator factory object
func NewCryptoCurrencyInfoHTTPFetcherCreator(client fetchers.HTTPClient) *CryptoCurrencyInfoHTTPFetcherCreator {
	return &CryptoCurrencyInfoHTTPFetcherCreator{client}
}

// Create creates a fetchers.CryptoCurrencyInfoFetcher instance for the given crypto-currency attached to the HTTP client specified in CryptoCurrencyInfoHttpFetcherCreator
func (creator *CryptoCurrencyInfoHTTPFetcherCreator) Create(symbol cryptoCurrencyTickerSymbol) (infoFetcher fetchers.CryptoCurrencyInfoFetcher, err error) {
	switch symbol {
	case btc:
		infoFetcher = fetchers.NewBlockchainInfoFetcher(creator.client)
	case eth:
		infoFetcher = fetchers.NewEtherscanInfoFetcher(creator.client)
	case bcc, dash, ltc, uno:
		currency := cryptoCurrencyMap[symbol]
		infoFetcher = fetchers.NewCryptoidInfoFetcher(currency, creator.client)
	}

	if infoFetcher == nil {
		err = fmt.Errorf("unknown crypto-currency %s", symbol)
	}

	return
}
