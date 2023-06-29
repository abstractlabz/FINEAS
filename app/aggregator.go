package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handleQuoteRequest)
	fmt.Println("Server listening on port 8080...")
	log.Println(http.ListenAndServe(":8080", nil))
}

func handleQuoteRequest(w http.ResponseWriter, r *http.Request) {

	type QueriedInfoAggregate struct {
		Ticker    string
		YTD_Info  string
		Fin_Info  string
		News_Info string
		Desc_Info string
	}

	// Create a new instance of QueriedInfoAggregate
	var queriedInfoAggregate QueriedInfoAggregate

	// Process query string parameters from the request URL
	queryParams := r.URL.Query()
	ticker := queryParams.Get("ticker")
	fmt.Println(ticker)

	// Return the entry url as the response
	w.Header().Set("Content-Type", "text/plain")

	// Get the financial information from the services
	// and return it as the response

	ytd_info := getFinancialInfo(ticker, "/ytd", "http://localhost:8081")
	queriedInfoAggregate.YTD_Info = ytd_info
	fmt.Println(ytd_info + "\n")

	fin_info := getFinancialInfo(ticker, "/fin", "http://localhost:8082")
	queriedInfoAggregate.Fin_Info = fin_info
	fmt.Println(fin_info + "\n")

	news_info := getFinancialInfo(ticker, "/news", "http://localhost:8083")
	queriedInfoAggregate.News_Info = news_info
	fmt.Println(news_info + "\n")

	desc_info := getFinancialInfo(ticker, "/desc", "http://localhost:8084")
	queriedInfoAggregate.Desc_Info = desc_info
	fmt.Println(desc_info + "\n")

}

func getFinancialInfo(ticker string, handlerID string, handlerURL string) string {

	base_url := handlerURL + handlerID

	// Construct the URL with query parameters
	url := base_url + "?" + "ticker=" + ticker

	// Send a GET request
	getResponse, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer getResponse.Body.Close()

	// Read the response body
	getResponseBody, err := ioutil.ReadAll(getResponse.Body)
	if err != nil {
		log.Println(err)
	}

	return string(getResponseBody)
}
