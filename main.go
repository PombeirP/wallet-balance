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

	// Load crypto-currency accounts
	currenciesConfig, err := loadConfigFromJSONFile("./config.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	results := make(chan *CryptoCurrencyBalanceReport, len(currenciesConfig))
	client := &http.Client{Timeout: 10 * time.Second}
	currencyInfoFetcherCreator := NewCryptoCurrencyInfoHTTPFetcherCreator(client)
	go fetchBalanceReports(currenciesConfig, currencyInfoFetcherCreator, 3, results)

	// Wait for results
	var reports []*CryptoCurrencyBalanceReport
	for _ = range currenciesConfig {
		reports = append(reports, <-results)
	}

	// Sort balances
	slice.Sort(reports, func(i, j int) bool {
		bi, bj := reports[i], reports[j]
		return bi.Balance*bi.UsdExchangeRate > bj.Balance*bj.UsdExchangeRate
	})

	printReports(reports)
}

func fetchBalanceReports(currenciesConfig []*cryptoBalanceCheckerConfig, currencyInfoFetcherCreator CryptoCurrencyInfoFetcherCreator, workerCount int, results chan<- *CryptoCurrencyBalanceReport) {
	jobs := make(chan *cryptoBalanceCheckerConfig, workerCount)

	// Define worker
	worker := func(jobs <-chan *cryptoBalanceCheckerConfig, results chan<- *CryptoCurrencyBalanceReport) {
		for j := range jobs {
			if infoFetcher, err := currencyInfoFetcherCreator.Create(j.Symbol); err == nil {
				go FetchInfoForCryptoCurrency(j, infoFetcher, results)
			} else {
				results <- nil
			}
		}
	}

	// Start up some workers
	for index := 0; index < cap(jobs); index++ {
		go worker(jobs, results)
	}

	// Wait for all jobs to complete
	for _, currencyConfig := range currenciesConfig {
		jobs <- currencyConfig
	}
}

func printReports(reports []*CryptoCurrencyBalanceReport) {
	// Calculate max symbol length for formatting
	var maxSymbolLength int
	for _, report := range reports {
		if length := len(report.Symbol); length > maxSymbolLength {
			maxSymbolLength = length
		}
	}

	// Print report
	totalUsdBalance := 0.
	usdColor := color.New(color.FgHiGreen).SprintFunc()
	cryptoColor := color.New(color.FgHiCyan).SprintFunc()
	errorColor := color.New(color.FgHiRed).SprintFunc()
	for _, report := range reports {
		if report.Error != nil {
			fmt.Fprintf(color.Output, "%s: %s\n", report.Symbol, errorColor(report.Error))
		} else {
			usdBalance := report.Balance * report.UsdExchangeRate
			totalUsdBalance += usdBalance

			cryptoBalanceString := fmt.Sprintf(fmt.Sprintf("%%%df", 13-len(report.Symbol)), report.Balance)
			cryptoTickerSymbolString := fmt.Sprintf(fmt.Sprintf("%%%ds", -maxSymbolLength), report.Symbol)

			fmt.Fprintf(color.Output, "%s balance: %s %s (in USD: %s, %s%s = %s)\n",
				report.Symbol,
				cryptoColor(cryptoBalanceString),
				cryptoTickerSymbolString,
				usdColor(fmt.Sprintf("%7.2f$", usdBalance)),
				cryptoColor("1"),
				cryptoTickerSymbolString,
				usdColor(fmt.Sprintf("%.2f$", report.UsdExchangeRate)))
		}
	}
	fmt.Println("------------------------------------------")
	fmt.Fprintf(color.Output, "USD balance: %s\n", usdColor(fmt.Sprintf("%.2f$", totalUsdBalance)))
}
