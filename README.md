# wallet-balance

Simple console Go project to retrieve balances for multiple cryptocurrency addresses

## Usage

- Copy `config.sample.json` to `config.json` and adapt it to your particular scenario. BTC, DASH and LTC are currently supported. The `chainz.cryptoid.info` API key is optional.go go 
- Run the program with `go build && ./wallet-balance`

## Sample output

```
BTC balance:   0.277827 BTC  (in USD: 3279.35$, 1BTC  = 11803.59$)
DASH balance:  0.079065 DASH (in USD:   58.71$, 1DASH = 742.61$)
LTC balance:   0.184422 LTC  (in USD:   18.98$, 1LTC  = 102.92$)
ETH balance:   0.000000 ETH  (in USD:    0.00$, 1ETH  = 457.23$)
------------------------------------------
USD balance: 3357.05$
```