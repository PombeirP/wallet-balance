package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type cryptoBalanceCheckerConfig struct {
	Symbol    string   `json:"symbol,omitempty"`
	Addresses []string `json:"addresses,omitempty"`
	APIKey    string   `json:"api_key,omitempty"`
}

func loadConfigFromJSON() []*CryptoBalanceChecker {
	raw, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var currencies []*cryptoBalanceCheckerConfig
	json.Unmarshal(raw, &currencies)

	c := make([]*CryptoBalanceChecker, len(currencies))
	for idx, checkerConfig := range currencies {
		c[idx] = NewCryptoBalanceChecker(checkerConfig.Symbol, checkerConfig.APIKey, checkerConfig.Addresses...)
	}
	return c
}
