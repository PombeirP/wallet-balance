package main

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCryptoCurrencyInfoFetcher struct {
	mock.Mock
}

func (m *MockCryptoCurrencyInfoFetcher) FetchBalance(addresses []string, apiKey string, balance *float64, err *error, done *sync.WaitGroup) {
	m.Called(addresses, apiKey, balance, err, done)
	done.Done()
	return
}

func (m *MockCryptoCurrencyInfoFetcher) FetchExchangeRate(apiKey string, targetCurrency string, exchangeRate *float64, err *error, done *sync.WaitGroup) {
	m.Called(apiKey, targetCurrency, exchangeRate, err, done)
	done.Done()
	return
}

func TestFetchInfoForCryptoCurrency(t *testing.T) {
	cases := []struct {
		name                    string
		symbol                  cryptoCurrencyTickerSymbol
		apiKey                  string
		addresses               []string
		returnedBalanceErr      error
		returnedExchangeRateErr error
		returnedBalance         float64
		returnedUsdExchangeRate float64
		returnedErrMessage      string
	}{
		{"BTC", btc, "random_api_key#1", []string{"a", "b"}, nil, nil, 1000., 99., ""},
		{"ETH", eth, "random_api_key#2", []string{"d"}, nil, nil, 50., 3., ""},
		{"balance error is propagated", eth, "random_api_key#2", []string{"d"}, errors.New("balance retrieval error"), nil, 0., 4., "Balance retrieval error"},
		{"exchange rate error is propagated", eth, "random_api_key#2", []string{"d"}, nil, errors.New("exchange rate retrieval error"), 0., 4., "Exchange rate retrieval error"},
	}

	for _, testCase := range cases {
		done := make(chan *CryptoCurrencyBalanceReport, 2)

		config := &cryptoBalanceCheckerConfig{testCase.symbol, testCase.addresses, testCase.apiKey}

		infoFetcherMock := new(MockCryptoCurrencyInfoFetcher)
		infoFetcherMock.On("FetchBalance", testCase.addresses, testCase.apiKey, mock.AnythingOfType("*float64"), mock.AnythingOfType("*error"), mock.AnythingOfType("chan<- bool")).Once().Run(func(args mock.Arguments) {
			*(args.Get(2).(*float64)) = testCase.returnedBalance
			*(args.Get(3).(*error)) = testCase.returnedBalanceErr
			args.Get(4).(chan<- bool) <- true
		})
		infoFetcherMock.On("FetchExchangeRate", testCase.apiKey, "usd", mock.AnythingOfType("*float64"), mock.AnythingOfType("*error"), mock.AnythingOfType("chan<- bool")).Once().Run(func(args mock.Arguments) {
			*(args.Get(2).(*float64)) = testCase.returnedUsdExchangeRate
			*(args.Get(3).(*error)) = testCase.returnedExchangeRateErr
			args.Get(4).(chan<- bool) <- true
		})

		FetchInfoForCryptoCurrency(config, infoFetcherMock, done)

		infoFetcherMock.AssertExpectations(t)

		report := <-done
		require.NotNil(t, report)
		require.Equal(t, testCase.symbol, report.Symbol)
		require.Equalf(t, testCase.returnedBalance, report.Balance, "Balance reported (%f) does not matched expected value (%f)", report.Balance, testCase.returnedBalance)
		require.Equalf(t, testCase.returnedUsdExchangeRate, report.UsdExchangeRate, "Exchange rate reported (%f) does not matched expected value (%f)", report.UsdExchangeRate, testCase.returnedUsdExchangeRate)
		if testCase.returnedErrMessage == "" {
			require.Nil(t, report.Error)
		} else {
			require.EqualErrorf(t, report.Error, testCase.returnedErrMessage, "Error reported (%s) does not matched expected value (%s)", report.Error, testCase.returnedErrMessage)
		}
	}
}
