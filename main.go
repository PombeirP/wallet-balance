package main

import (
	"fmt"
	"net/http"
)

func main() {
	client := &http.Client{}

	// Load balance checkers for each crypto-currency
	balanceCheckers := loadConfigFromJSON()
	done := make([]chan bool, len(balanceCheckers))

	// Kick off asynchronous balance check for each crypto-currency
	for idx, checker := range balanceCheckers {
		done[idx] = make(chan bool)

		go checker.GetAddressBalances(client, done[idx])
	}

	// Wait for result and check for errors
	var maxSymbolLength int
	for idx, checker := range balanceCheckers {
		if <-done[idx]; checker.Error != nil {
			fmt.Println(checker.Error)
			return
		}

		if length := len(checker.Symbol); length > maxSymbolLength {
			maxSymbolLength = length
		}
	}

	// Print out balances
	totalUsdBalance := 0.
	for _, checker := range balanceCheckers {
		format := fmt.Sprintf("%s balance: %%%df %%%ds (in USD: %%7.2f$)\n", checker.Symbol, 13-len(checker.Symbol), -maxSymbolLength)
		usdBalance := checker.Balance * checker.UsdExchangeRate
		totalUsdBalance += usdBalance
		fmt.Printf(format, checker.Balance, checker.Symbol, usdBalance)
	}
	fmt.Println("------------------------------------------")
	fmt.Printf("USD balance: %.2f$\n", totalUsdBalance)
}
