package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/bradfitz/slice"
	"github.com/fatih/color"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	fmt.Println("Fetching balances...")

	// Load balance checkers for each crypto-currency
	balanceCheckers, err := loadConfigFromJSONFile("./config.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

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
	slice.Sort(balanceCheckers, func(i, j int) bool {
		bi, bj := balanceCheckers[i], balanceCheckers[j]
		return bi.Balance*bi.UsdExchangeRate > bj.Balance*bj.UsdExchangeRate
	})

	totalUsdBalance := 0.
	usdColor := color.New(color.FgHiGreen).SprintFunc()
	cryptoColor := color.New(color.FgHiCyan).SprintFunc()
	errorColor := color.New(color.FgHiRed).SprintFunc()
	for _, checker := range balanceCheckers {
		if checker.Error != nil {
			fmt.Fprintf(color.Output, "%s: %s\n", checker.Symbol, errorColor(checker.Error))
		} else {
			usdBalance := checker.Balance * checker.UsdExchangeRate
			totalUsdBalance += usdBalance

			cryptoBalanceString := fmt.Sprintf(fmt.Sprintf("%%%df", 13-len(checker.Symbol)), checker.Balance)
			cryptoTickerSymbolString := fmt.Sprintf(fmt.Sprintf("%%%ds", -maxSymbolLength), checker.Symbol)

			fmt.Fprintf(color.Output, "%s balance: %s %s (in USD: %s, %s%s = %s)\n",
				checker.Symbol,
				cryptoColor(cryptoBalanceString),
				cryptoTickerSymbolString,
				usdColor(fmt.Sprintf("%7.2f$", usdBalance)),
				cryptoColor("1"),
				cryptoTickerSymbolString,
				usdColor(fmt.Sprintf("%.2f$", checker.UsdExchangeRate)))
		}
	}
	fmt.Println("------------------------------------------")
	fmt.Fprintf(color.Output, "USD balance: %s\n", usdColor(fmt.Sprintf("%.2f$", totalUsdBalance)))
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
