package main

import (
	"fmt"
	"net/http"
)

func main() {
	client := &http.Client{}

	// Load balance checkers for each crypto-currency
	balanceCheckers := loadConfigFromJSON()
	done := make(chan *CryptoBalanceChecker)

	// Kick off asynchronous balance check for each crypto-currency
	for _, checker := range balanceCheckers {
		go checker.GetAddressBalances(client, done)
	}

	// Wait for result and check for errors
	var maxSymbolLength int
	for index := 0; index < len(balanceCheckers); index++ {
		completedCheck := <-done
		if completedCheck.Error != nil {
			fmt.Println(completedCheck.Error)
			return
		}

		if length := len(completedCheck.Symbol); length > maxSymbolLength {
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
