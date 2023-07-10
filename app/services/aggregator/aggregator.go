package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
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

	queriedInfoAggregate.Ticker = ticker
	ytd_info := getFinancialInfo(ticker, "/ytd", "http://localhost:8081")
	queriedInfoAggregate.YTD_Info = ytd_info

	fin_info := getFinancialInfo(ticker, "/fin", "http://localhost:8082")
	queriedInfoAggregate.Fin_Info = fin_info

	news_info := getFinancialInfo(ticker, "/news", "http://localhost:8083")
	queriedInfoAggregate.News_Info = news_info

	desc_info := getFinancialInfo(ticker, "/desc", "http://localhost:8084")
	queriedInfoAggregate.Desc_Info = desc_info

	// Marshal the queriedInfoAggregate struct into json
	queriedInfoAggregate_json, err := json.Marshal(queriedInfoAggregate)
	if err != nil {
		log.Println(err)
	}

	template := "You are an AI who is tasked to summarize financial information for this company ticker. Give me a summary on the financial health of this company based on the following data. This is the data:"
	prompt_inference := getPromptInference(string(queriedInfoAggregate_json), template, "/", "http://localhost:8086")
	// Return the queriedInfoAggregate_json as the response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(prompt_inference))

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

func getPromptInference(prompt string, template string, handlerID string, handlerURL string) string {

	base_url := handlerURL + handlerID

	url := base_url + "?" + "prompt=" + urlConverter(template+prompt)

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

func urlConverter(_url string) string {

	// Construct the URL with query parameters
	encoded_prompt := url.QueryEscape(_url)
	encoded_prompt = strings.ReplaceAll(encoded_prompt, "+", "%20")
	encoded_prompt = strings.ReplaceAll(encoded_prompt, "%2F", "/")
	encoded_prompt = strings.ReplaceAll(encoded_prompt, "%3A", ":")

	return encoded_prompt
}
