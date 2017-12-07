package fetchers

// CryptocurrencyBalanceFetcher defines the interface for fetching a crypto-currency balance
type CryptocurrencyBalanceFetcher interface {
	FetchBalance(addresses []string, apiKey string, balance *float64, err *error, done chan<- bool)
}

// CryptocurrencyExchangeRateFetcher defines the interface for fetching a crypto-currency exchange rate
type CryptocurrencyExchangeRateFetcher interface {
	FetchExchangeRate(apiKey string, targetCurrency string, exchangeRate *float64, err *error, done chan<- bool)
}

// CryptocurrencyInfoFetcher defines the interface for fetching the balance and exchange rate of a crypto-currency
type CryptocurrencyInfoFetcher interface {
	CryptocurrencyBalanceFetcher
	CryptocurrencyExchangeRateFetcher
}
