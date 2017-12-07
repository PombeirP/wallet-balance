package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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
	const targetCurrency = "usd"

	balancesFetched := make(chan bool)
	exchangeRateFetched := make(chan bool)

	switch checker.Symbol {
	case btc:
		go checker.getBlockchainAddressBalances(client, balancesFetched)
		go checker.getBlockchainExchangeRate(client, targetCurrency, exchangeRateFetched)
	case eth:
		go checker.getEtherscanAddressBalances(client, balancesFetched)
		go checker.getEtherscanExchangeRate(client, targetCurrency, exchangeRateFetched)
	case bcc, dash, ltc, uno:
		currency := cryptoCurrencyMap[checker.Symbol]
		go checker.getCryptoidAddressBalances(client, currency, balancesFetched)
		go checker.getCryptoidExchangeRate(client, currency, targetCurrency, exchangeRateFetched)
	default:
		checker.Error = fmt.Errorf("Unknown crypto-currency %s", checker.Symbol)
		done <- checker
		return
	}

	<-exchangeRateFetched
	<-balancesFetched
	done <- checker
}

type etherscanResponseHeader struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// getEtherscanAddressBalances retrieves the aggregate balances for the previously provided addresses
func (checker *CryptoBalanceChecker) getEtherscanAddressBalances(client *http.Client, done chan<- bool) {
	checker.Balance = 0.
	checker.Error = nil

	type etherscanAccountBalanceResult struct {
		Account string `json:"account,omitempty"`
		Balance string `json:"balance,omitempty"`
	}

	type etherscanAccountBalanceResponse struct {
		etherscanResponseHeader
		Result []*etherscanAccountBalanceResult `json:"result,omitempty"`
	}

	url := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=balancemulti&address=%s&tag=latest", strings.Join(checker.Addresses, ","))
	response := &etherscanAccountBalanceResponse{}
	responseReadyChan := make(chan bool)
	errorsChan := make(chan error)

	go fetchJSONResponse(client, url, response, responseReadyChan, errorsChan)

	select {
	case err := <-errorsChan:
		checker.Error = err
	case <-responseReadyChan:
		for _, responseEntry := range response.Result {
			balance, err := strconv.ParseFloat(responseEntry.Balance, 64)
			if err != nil {
				checker.Error = err
			} else {
				checker.Balance += balance
			}
		}
	}

	done <- true

	return
}

// getEtherscanExchangeRate retrieves the exchange rate for BTC in `targetCurrency`
func (checker *CryptoBalanceChecker) getEtherscanExchangeRate(client *http.Client, targetCurrency string, done chan<- bool) {
	checker.UsdExchangeRate = 0.

	if targetCurrency != "usd" {
		checker.Error = fmt.Errorf("%s is not supported as target currency for ETH, only USD at the moment", targetCurrency)
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

	url := fmt.Sprintf("https://api.etherscan.io/api?module=stats&action=ethprice&apikey=%s", checker.APIKey)
	go fetchJSONResponse(client, url, response, responseReadyChan, errorsChan)

	select {
	case err := <-errorsChan:
		checker.Error = err
	case <-responseReadyChan:
		exchangeRate, err := strconv.ParseFloat(response.Result.ETHUSD, 64)
		if err != nil {
			checker.Error = err
		} else {
			checker.UsdExchangeRate = exchangeRate
		}
	}

	done <- true
}

// GetCryptoidAddressBalances retrieves the aggregate balances for the previously provided addresses
func (checker *CryptoBalanceChecker) getCryptoidAddressBalances(client *http.Client, currency string, done chan<- bool) {
	checker.Error = nil
	checker.Balance = 0.

	balances := make(chan float64)
	errorsChan := make(chan error)

	for _, address := range checker.Addresses {
		url := fmt.Sprintf("https://chainz.cryptoid.info/%s/api.dws?q=getbalance&key=%s&a=%s", currency, checker.APIKey, address)
		go fetchValueFromURL(client, url, balances, errorsChan)
	}

	for _ = range checker.Addresses {
		select {
		case err := <-errorsChan:
			checker.Error = err
		case balance := <-balances:
			checker.Balance += balance
		}
	}

	done <- true
}

// getCryptoidExchangeRate retrieves the exchange rate for `currency` in `targetCurrency`
func (checker *CryptoBalanceChecker) getCryptoidExchangeRate(client *http.Client, currency string, targetCurrency string, done chan<- bool) {
	checker.UsdExchangeRate = 0.

	exchangeRates := make(chan float64)
	errorsChan := make(chan error)

	url := fmt.Sprintf("https://chainz.cryptoid.info/%s/api.dws?q=ticker.%s&key=%s", currency, targetCurrency, checker.APIKey)
	go fetchValueFromURL(client, url, exchangeRates, errorsChan)

	select {
	case err := <-errorsChan:
		checker.Error = err
		break
	case exchangeRate := <-exchangeRates:
		checker.UsdExchangeRate = exchangeRate
	}

	done <- true
}

// getBlockchainAddressBalances retrieves the aggregate balances for the previously provided addresses
func (checker *CryptoBalanceChecker) getBlockchainAddressBalances(client *http.Client, done chan<- bool) {
	balances := make(chan float64)
	errorsChan := make(chan error)

	url := fmt.Sprintf("https://blockchain.info/q/addressbalance/%s", strings.Join(checker.Addresses, "%7C" /*|*/))
	go fetchValueFromURL(client, url, balances, errorsChan)

	select {
	case err := <-errorsChan:
		checker.Error = err
		break
	case balance := <-balances:
		checker.Balance = balance / 100000000.
	}

	done <- true
}

// getBlockchainExchangeRate retrieves the exchange rate for BTC in `targetCurrency`
func (checker *CryptoBalanceChecker) getBlockchainExchangeRate(client *http.Client, targetCurrency string, done chan<- bool) {
	checker.UsdExchangeRate = 0.

	exchangeRates := make(chan float64)
	errorsChan := make(chan error)

	url := fmt.Sprintf("https://blockchain.info/tobtc?currency=%s&value=1", targetCurrency)
	go fetchValueFromURL(client, url, exchangeRates, errorsChan)

	select {
	case err := <-errorsChan:
		checker.Error = err
		break
	case exchangeRate := <-exchangeRates:
		checker.UsdExchangeRate = 1. / exchangeRate
	}

	done <- true
}

func fetchValueFromURL(client *http.Client, url string, resultChan chan<- float64, errorsChan chan<- error) {
	resp, err := client.Get(url)
	if err != nil {
		errorsChan <- err
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorsChan <- err
		return
	}

	bodyString := string(body)
	if resp.StatusCode >= 300 {
		if len(bodyString) > 0 {
			errorsChan <- errors.New(bodyString)
		} else {
			errorsChan <- errors.New(resp.Status)
		}

		return
	}

	value, err := strconv.ParseFloat(bodyString, 64)
	if err != nil {
		errorsChan <- err
		return
	}

	resultChan <- value
}

func fetchJSONResponse(client *http.Client, url string, response interface{}, responseReadyChan chan<- bool, errorsChan chan<- error) {
	resp, err := client.Get(url)
	if err != nil {
		errorsChan <- err
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorsChan <- err
		return
	}

	json.Unmarshal([]byte(body), response)

	var untypedResponse map[string]interface{}
	err = json.Unmarshal([]byte(body), &untypedResponse)
	if err != nil {
		errorsChan <- err
		return
	}

	if untypedResponse["status"].(string) != "1" {
		errorsChan <- errors.New(untypedResponse["message"].(string))
		return
	}

	responseReadyChan <- true
}
