package main

import (
	"encoding/json"
	"io"
	"os"
)

type cryptoBalanceCheckerConfig struct {
	Symbol    cryptoCurrencyTickerSymbol `json:"symbol,omitempty"`
	Addresses []string                   `json:"addresses,omitempty"`
	APIKey    string                     `json:"api_key,omitempty"`
}

func loadConfigFromJSONFile(path string) ([]*CryptoBalanceChecker, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	raw := make([]byte, info.Size())
	_, err = io.ReadFull(file, raw)
	if err != nil {
		return nil, err
	}

	return loadConfigFromJSON(raw), nil
}

func loadConfigFromJSON(rawJSON []byte) (checker []*CryptoBalanceChecker) {
	var currencies []*cryptoBalanceCheckerConfig
	json.Unmarshal(rawJSON, &currencies)

	checker = make([]*CryptoBalanceChecker, len(currencies))
	for idx, checkerConfig := range currencies {
		checker[idx] = NewCryptoBalanceChecker(checkerConfig.Symbol, checkerConfig.APIKey, checkerConfig.Addresses...)
	}

	return checker
}
