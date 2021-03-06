package fetchers

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

const (
	wei = 1000000000000000000. // 10^18
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
func (fetcher *EtherscanInfoFetcher) FetchBalance(addresses []string, apiKey string, balance *float64, err *error, done *sync.WaitGroup) {
	defer done.Done()

	*balance = 0.

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

	*err = fetcher.apiFetcher.Fetch(url, response)

	if *err == nil {
		for _, responseEntry := range response.Result {
			partialBalance, errParse := strconv.ParseFloat(responseEntry.Balance, 64)
			if errParse == nil {
				*balance += partialBalance / wei
			} else {
				*err = errParse
				break
			}
		}
	}
}

// FetchExchangeRate retrieves the exchange rate for ETH in `targetCurrency` from https://api.etherscan.io/
func (fetcher *EtherscanInfoFetcher) FetchExchangeRate(apiKey string, targetCurrency string, exchangeRate *float64, err *error, done *sync.WaitGroup) {
	defer done.Done()

	*exchangeRate = 0.

	if targetCurrency != "usd" {
		*err = fmt.Errorf("%s is not supported as target currency for ETH, only USD at the moment", targetCurrency)
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

	url := fmt.Sprintf("https://api.etherscan.io/api?module=stats&action=ethprice&apikey=%s", apiKey)
	*err = fetcher.apiFetcher.Fetch(url, response)

	if *err == nil {
		_exchangeRate, _err := strconv.ParseFloat(response.Result.ETHUSD, 64)
		if _err != nil {
			*err = _err
		} else {
			*exchangeRate = _exchangeRate
		}
	}
}
