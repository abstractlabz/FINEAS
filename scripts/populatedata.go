package main

import (
	"fmt"
	"net/http"
)

func main() {
	stockTickers := []string{
		// S&P 500
		"AAPL",
		"MSFT",
		"AMZN",
		"GOOGL",
		"TSLA",
		"META",
		"JPM",
		"V",
		"MA",
		"NVDA",
		"UNH",
		"HD",
		"WMT",
		"P&G",
		"VICI",
		"DIS",
		"COST",
		"CVX",
		"XOM",
		"MCD",
		"NKE",
		"JNJ",
		"BAC",
		"MRK",
		"DOW",
		"CSCO",
		"TXN",
		"PEP",
		"KO",
		"INTC",
		"IBM",
		"CRM",
		"MMM",
		"LLY",
		"MCDY",
		"ADM",
		"TMO",
		"DHR",
		"CAT",
		"UPS",
		"VZ",
		"NEE",
		"PFE",
		"TGT",
		"LMT",
		"AMGN",
		"UNP",
		"HON",
		"CHTR",
		"LHX",
		"BKNG",
		"NFLX",
		"PYPL",
		"ADBE",
		"NVRO",
		"MRVL",
		"QCOM",
		"ASML",
		"TSM",
		"MU",
		"TXMD",
		"AVGO",
		"MAR",
		"KLAC",
		"SWKS",
		"SNPS",
		"ASAN",
		"ALGN",
		"MDB",
		"CRWD",
		"ZM",
		"SPOT",
		"DOCU",
		"FSLY",
		"TWTR",
		"SNAP",
		"PINS",
		"ROKU",
		"SQ",

		// Dow Jones Industrial Average
		"MMM",
		"AXP",
		"AMGN",
		"AAPL",
		"BA",
		"CAT",
		"CSCO",
		"CVX",
		"DOW",
		"GS",
		"HD",
		"HON",
		"IBM",
		"INTC",
		"JPM",
		"JNJ",
		"KO",
		"MCD",
		"MRK",
		"MSFT",
		"NKE",
		"P&G",
		"TRV",
		"UNH",
		"V",
		"VZ",
		"WMT",
		"XOM",

		// Nasdaq 100
		"AAPL",
		"MSFT",
		"AMZN",
		"GOOGL",
		"TSLA",
		"META",
		"NVDA",
		"GOOG",
		"ADBE",
		"PEP",
		"INTC",
		"PYPL",
		"COST",
		"TXN",
		"QCOM",
		"AVGO",
		"INTU",
		"ZM",
		"ASML",
		"TSM",
		"ROKU",
		"NFLX",
		"AMD",
		"JD",
		"MELI",
		"PINS",
		"SQ",
		"DOCU",
		"SE",
		"SPOT",
		"SNOW",
		"TWTR",
		"ROST",
		"OKTA",
		"ZS",
		"COUP",
		"PANW",
		"BILL",
		"NET",
		"FSLY",
		"TWLO",
		"CRWD",
		"VRTX",
		"HUBS",
		"MRVL",
		"CFLT",
		"ABNB",
		"DASH",
		"DOOR",
		"COIN",
		"UPST",
		"FVRR",
		"DATADOG",
		"TTD",
	}

	// iterate through the list of stock tickers
	for _, ticker := range stockTickers {
		// Create a new HTTP client
		client := &http.Client{}

		// Construct the URL with query parameters
		base_url := "http://localhost:8080/"
		url := base_url + "?" + "ticker=" + ticker

		// Create a GET request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}
		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println(ticker)
	}

}
