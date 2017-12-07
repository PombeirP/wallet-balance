package main

import (
	"fmt"
	"net/http"

	"github.com/PombeirP/wallet-balance/fetchers"
)

// cryptoCurrencyTickerSymbol represents the ticker symbol for a crypto-currency
type cryptoCurrencyTickerSymbol string

const (
	btc  cryptoCurrencyTickerSymbol = "BTC"
	eth  cryptoCurrencyTickerSymbol = "ETH"
	ltc  cryptoCurrencyTickerSymbol = "LTC"
	dash cryptoCurrencyTickerSymbol = "DASH"
	uno  cryptoCurrencyTickerSymbol = "UNO"
	bcc  cryptoCurrencyTickerSymbol = "BCC"
)

// cryptoCurrencyMap maps cryptoCurrencyTickerSymbol values to internal symbols used by external APIs
var cryptoCurrencyMap map[cryptoCurrencyTickerSymbol]string

// CryptoBalanceChecker provides functionality to check for the aggregate balance of crypto-currency addresses
type CryptoBalanceChecker struct {
	Symbol          cryptoCurrencyTickerSymbol
	Addresses       []string
	APIKey          string
	UsdExchangeRate float64

	Balance float64
	Error   error
}

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

// NewCryptoBalanceChecker creates a crypto-currency balance checker instance for given crypto-currency addresses
func NewCryptoBalanceChecker(symbol cryptoCurrencyTickerSymbol, APIKey string, addresses ...string) *CryptoBalanceChecker {
	return &CryptoBalanceChecker{symbol, addresses, APIKey, 0., 0., nil}
}

// GetAddressBalances retrieves the aggregate balances for the previously provided addresses
func (checker *CryptoBalanceChecker) GetAddressBalances(client *http.Client, done chan<- *CryptoBalanceChecker) {
	balancesFetched := make(chan bool)
	exchangeRateFetched := make(chan bool)

	var infoFetcher fetchers.CryptocurrencyInfoFetcher
	switch checker.Symbol {
	case btc:
		infoFetcher = fetchers.NewBlockchainInfoFetcher(client)
	case eth:
		infoFetcher = fetchers.NewEtherscanInfoFetcher(client)
	case bcc, dash, ltc, uno:
		currency := cryptoCurrencyMap[checker.Symbol]
		infoFetcher = fetchers.NewCryptoidInfoFetcher(currency, client)
	}

	if infoFetcher != nil {
		go infoFetcher.FetchBalance(checker.Addresses, checker.APIKey, &checker.Balance, &checker.Error, balancesFetched)
		go infoFetcher.FetchExchangeRate(checker.APIKey, "usd", &checker.UsdExchangeRate, &checker.Error, exchangeRateFetched)

		<-exchangeRateFetched
		<-balancesFetched
	} else {
		checker.Error = fmt.Errorf("Unknown crypto-currency %s", checker.Symbol)
	}

	done <- checker
}
