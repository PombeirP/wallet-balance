package fetchers

import (
	"fmt"
	"strconv"
	"strings"
)

// EtherscanInfoFetcher fetches the balance and exchange rate of Ethereum on https://api.etherscan.io/
type EtherscanInfoFetcher struct {
	apiFetcher JSONFetcher
}

// NewEtherscanInfoFetcher creates an instance of EtherscanInfoFetcher from an HTTP client instance
func NewEtherscanInfoFetcher(client HTTPClient) *EtherscanInfoFetcher {
	apiFetcher := NewEtherscanJSONFetcher(client)
	return &EtherscanInfoFetcher{apiFetcher}
}

type etherscanResponseHeader struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// FetchBalance retrieves the balance for the specified addresses from https://api.etherscan.io/
func (fetcher *EtherscanInfoFetcher) FetchBalance(addresses []string, apiKey string, balance *float64, err *error, done chan<- bool) {
	*balance = 0.
	*err = nil

	type etherscanAccountBalanceResult struct {
		Account string `json:"account,omitempty"`
		Balance string `json:"balance,omitempty"`
	}

	type etherscanAccountBalanceResponse struct {
		etherscanResponseHeader
		Result []*etherscanAccountBalanceResult `json:"result,omitempty"`
	}

	url := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=balancemulti&address=%s&tag=latest", strings.Join(addresses, ","))
	response := &etherscanAccountBalanceResponse{}
	responseReadyChan := make(chan bool)
	errorsChan := make(chan error)

	go fetcher.apiFetcher.Fetch(url, response, responseReadyChan, errorsChan)

	select {
	case *err = <-errorsChan:
	case <-responseReadyChan:
		for _, responseEntry := range response.Result {
			partialBalance, err := strconv.ParseFloat(responseEntry.Balance, 64)
			if err == nil {
				*balance += partialBalance
			}
		}
	}

	done <- true

	return
}

// FetchExchangeRate retrieves the exchange rate for ETH in `targetCurrency` from https://api.etherscan.io/
func (fetcher *EtherscanInfoFetcher) FetchExchangeRate(apiKey string, targetCurrency string, exchangeRate *float64, err *error, done chan<- bool) {
	*exchangeRate = 0.

	if targetCurrency != "usd" {
		*err = fmt.Errorf("%s is not supported as target currency for ETH, only USD at the moment", targetCurrency)
		done <- true
		return
	}

	type etherscanEthPriceResult struct {
		ETHUSD string `json:"ethusd,omitempty"`
	}

	type etherscanEthPriceResponse struct {
		etherscanResponseHeader
		Result *etherscanEthPriceResult `json:"result,omitempty"`
	}
	response := &etherscanEthPriceResponse{}
	responseReadyChan := make(chan bool)
	errorsChan := make(chan error)

	url := fmt.Sprintf("https://api.etherscan.io/api?module=stats&action=ethprice&apikey=%s", apiKey)
	go fetcher.apiFetcher.Fetch(url, response, responseReadyChan, errorsChan)

	select {
	case *err = <-errorsChan:
	case <-responseReadyChan:
		_exchangeRate, _err := strconv.ParseFloat(response.Result.ETHUSD, 64)
		if _err != nil {
			*err = _err
		} else {
			*exchangeRate = _exchangeRate
		}
	}

	done <- true
}
