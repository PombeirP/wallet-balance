package fetchers

import "sync"

// CryptoCurrencyBalanceFetcher defines the interface for fetching a crypto-currency balance
type CryptoCurrencyBalanceFetcher interface {
	FetchBalance(addresses []string, apiKey string, balance *float64, err *error, done *sync.WaitGroup)
}

// CryptoCurrencyExchangeRateFetcher defines the interface for fetching a crypto-currency exchange rate
type CryptoCurrencyExchangeRateFetcher interface {
	FetchExchangeRate(apiKey string, targetCurrency string, exchangeRate *float64, err *error, done *sync.WaitGroup)
}

// CryptoCurrencyInfoFetcher defines the interface for fetching the balance and exchange rate of a crypto-currency
type CryptoCurrencyInfoFetcher interface {
	CryptoCurrencyBalanceFetcher
	CryptoCurrencyExchangeRateFetcher
}
