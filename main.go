package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Fetching balances...")

	// Load balance checkers for each crypto-currency
	balanceCheckers := loadConfigFromJSON()

	done := make(chan bool)
	go fetchBalances(balanceCheckers, 3, done)

	// Calculate max symbol length for formatting
	var maxSymbolLength int
	for _, checker := range balanceCheckers {
		if length := len(checker.Symbol); length > maxSymbolLength {
			maxSymbolLength = length
		}
	}

	<-done

	// Print out balances
	totalUsdBalance := 0.
	for _, checker := range balanceCheckers {
		if checker.Error != nil {
			fmt.Printf("%s: %s\n", checker.Symbol, checker.Error)
		} else {
			format := fmt.Sprintf("%s balance: %%%df %%%ds (in USD: %%7.2f$, 1%%%ds = %%.2f$)\n", checker.Symbol, 13-len(checker.Symbol), -maxSymbolLength, -maxSymbolLength)
			usdBalance := checker.Balance * checker.UsdExchangeRate
			totalUsdBalance += usdBalance
			fmt.Printf(format, checker.Balance, checker.Symbol, usdBalance, checker.Symbol, checker.UsdExchangeRate)
		}
	}
	fmt.Println("------------------------------------------")
	fmt.Printf("USD balance: %.2f$\n", totalUsdBalance)
}

func fetchBalances(balanceCheckers []*CryptoBalanceChecker, workerCount int, done chan<- bool) {
	client := &http.Client{Timeout: 10 * time.Second}
	jobs := make(chan *CryptoBalanceChecker, workerCount)
	results := make(chan *CryptoBalanceChecker, len(balanceCheckers))

	// Start up some workers
	worker := func(jobs <-chan *CryptoBalanceChecker, results chan<- *CryptoBalanceChecker) {
		for j := range jobs {
			j.GetAddressBalances(client, results)
		}
	}
	for index := 0; index < cap(jobs); index++ {
		go worker(jobs, results)
	}

	// Kick off a job for each crypto-currency check
	for _, checker := range balanceCheckers {
		jobs <- checker
	}
	close(jobs)

	// Wait for results
	for _ = range balanceCheckers {
		<-results
	}

	done <- true
}
