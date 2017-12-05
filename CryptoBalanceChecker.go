package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// CryptoBalanceChecker provides functionality to check for the aggregate balance of crypto-currency addresses
type CryptoBalanceChecker struct {
	Symbol          string
	Addresses       []string
	APIKey          string
	UsdExchangeRate float64

	Balance float64
	Error   error
}

// NewCryptoBalanceChecker creates a crypto-currency balance checker instance for given crypto-currency addresses
func NewCryptoBalanceChecker(symbol string, APIKey string, addresses ...string) *CryptoBalanceChecker {
	return &CryptoBalanceChecker{symbol, addresses, APIKey, 0., 0., nil}
}

// GetAddressBalances retrieves the aggregate balances for the previously provided addresses
func (checker *CryptoBalanceChecker) GetAddressBalances(client *http.Client, done chan<- *CryptoBalanceChecker) {
	const targetCurrency = "usd"

	balancesFetched := make(chan bool)
	exchangeRateFetched := make(chan bool)

	switch checker.Symbol {
	case "BTC":
		go checker.getBlockchainAddressBalances(client, balancesFetched)
		go checker.getBlockchainExchangeRate(client, targetCurrency, exchangeRateFetched)
	case "ETH":
		go checker.getEtherscanAddressBalances(client, balancesFetched)
		go checker.getEtherscanExchangeRate(client, targetCurrency, exchangeRateFetched)
	case "BCC", "DASH", "LTC", "UNO":
		currency := strings.ToLower(checker.Symbol)
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

// getBlockchainAddressBalances retrieves the aggregate balances for the previously provided addresses
func (checker *CryptoBalanceChecker) getBlockchainAddressBalances(client *http.Client, done chan<- bool) {
	balances := make(chan float64)
	errors := make(chan error)

	url := fmt.Sprintf("https://blockchain.info/q/addressbalance/%s", strings.Join(checker.Addresses, "%7C" /*|*/))
	go fetchValueFromURL(client, url, balances, errors)

	select {
	case err := <-errors:
		checker.Error = err
		break
	case balance := <-balances:
		checker.Balance = balance / 100000000.
	}

	done <- true

	return
}

// getEtherscanAddressBalances retrieves the aggregate balances for the previously provided addresses
func (checker *CryptoBalanceChecker) getEtherscanAddressBalances(client *http.Client, done chan<- bool) {
	checker.Balance = 0.
	checker.Error = nil

	url := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=balancemulti&address=%s&tag=latest", strings.Join(checker.Addresses, ","))
	resp, err := client.Get(url)
	if err != nil {
		checker.Error = err
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		checker.Error = err
		return
	}

	type etherscanAccountBalance struct {
		Account string `json:"account,omitempty"`
		Balance string `json:"balance,omitempty"`
	}

	type etherscanResponse struct {
		Status  string                     `json:"status,omitempty"`
		Message string                     `json:"message,omitempty"`
		Result  []*etherscanAccountBalance `json:"result,omitempty"`
	}

	var response etherscanResponse
	json.Unmarshal([]byte(body), &response)

	if response.Status != "1" {
		checker.Error = fmt.Errorf(response.Message)
		return
	}

	for _, responseEntry := range response.Result {
		balance, err := strconv.ParseFloat(responseEntry.Balance, 64)
		if err != nil {
			checker.Error = err
		} else {
			checker.Balance += balance
		}
	}

	done <- true

	return
}

// getEtherscanExchangeRate retrieves the exchange rate for BTC in `targetCurrency`
func (checker *CryptoBalanceChecker) getEtherscanExchangeRate(client *http.Client, targetCurrency string, done chan<- bool) {
	checker.UsdExchangeRate = 0.

	if targetCurrency != "usd" {
		checker.Error = fmt.Errorf("%s is not supported as target currency for ETH, only USD at the moment")
		done <- true
		return
	}

	url := fmt.Sprintf("https://api.etherscan.io/api?module=stats&action=ethprice&apikey=%s", checker.APIKey)
	resp, err := client.Get(url)
	if err != nil {
		checker.Error = err
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		checker.Error = err
		return
	}

	type etherscanEthPriceResult struct {
		ETHUSD string `json:"ethusd,omitempty"`
	}

	type etherscanResponse struct {
		Status  string                   `json:"status,omitempty"`
		Message string                   `json:"message,omitempty"`
		Result  *etherscanEthPriceResult `json:"result,omitempty"`
	}

	var response etherscanResponse
	json.Unmarshal([]byte(body), &response)

	if response.Status != "1" {
		checker.Error = fmt.Errorf(response.Message)
		return
	}

	exchangeRate, err := strconv.ParseFloat(response.Result.ETHUSD, 64)
	if err != nil {
		checker.Error = err
	} else {
		checker.UsdExchangeRate = exchangeRate
	}

	done <- true
}

// GetCryptoidAddressBalances retrieves the aggregate balances for the previously provided addresses
func (checker *CryptoBalanceChecker) getCryptoidAddressBalances(client *http.Client, currency string, done chan<- bool) {
	checker.Error = nil
	checker.Balance = 0.

	balances := make(chan float64)
	errors := make(chan error)

	for _, address := range checker.Addresses {
		url := fmt.Sprintf("https://chainz.cryptoid.info/%s/api.dws?q=getbalance&key=%s&a=%s", currency, checker.APIKey, address)
		go fetchValueFromURL(client, url, balances, errors)
	}

	for _ = range checker.Addresses {
		select {
		case err := <-errors:
			checker.Error = err
		case balance := <-balances:
			checker.Balance += balance
		}
	}

	done <- true

	return
}

// getCryptoidExchangeRate retrieves the exchange rate for `currency` in `targetCurrency`
func (checker *CryptoBalanceChecker) getCryptoidExchangeRate(client *http.Client, currency string, targetCurrency string, done chan<- bool) {
	checker.UsdExchangeRate = 0.

	exchangeRates := make(chan float64)
	errors := make(chan error)

	url := fmt.Sprintf("https://chainz.cryptoid.info/%s/api.dws?q=ticker.%s&key=%s", currency, targetCurrency, checker.APIKey)
	go fetchValueFromURL(client, url, exchangeRates, errors)

	select {
	case err := <-errors:
		checker.Error = err
		break
	case exchangeRate := <-exchangeRates:
		checker.UsdExchangeRate = exchangeRate
	}

	done <- true

	return
}

// getBlockchainExchangeRate retrieves the exchange rate for BTC in `targetCurrency`
func (checker *CryptoBalanceChecker) getBlockchainExchangeRate(client *http.Client, targetCurrency string, done chan<- bool) {
	checker.UsdExchangeRate = 0.

	exchangeRates := make(chan float64)
	errors := make(chan error)

	url := fmt.Sprintf("https://blockchain.info/tobtc?currency=%s&value=1", targetCurrency)
	go fetchValueFromURL(client, url, exchangeRates, errors)

	select {
	case err := <-errors:
		checker.Error = err
		break
	case exchangeRate := <-exchangeRates:
		checker.UsdExchangeRate = 1. / exchangeRate
	}

	done <- true

	return
}

func fetchValueFromURL(client *http.Client, url string, result chan<- float64, errors chan<- error) {
	resp, err := client.Get(url)
	if err != nil {
		errors <- err
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errors <- err
		return
	}

	value, err := strconv.ParseFloat(string(body), 64)
	if err != nil {
		errors <- err
		return
	}

	result <- value
	return
}
