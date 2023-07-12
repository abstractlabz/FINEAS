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

	type PromptInference struct {
		Stock_Performance string
		Financial_Health  string
		News_Summary      string
		Company_Desc      string
	}

	var promptInference PromptInference

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

	/* Marshal the queriedInfoAggregate struct into json
	queriedInfoAggregate_json, err := json.Marshal(queriedInfoAggregate)
	if err != nil {
		log.Println(err)
	}
	*/

	// stock perfomance
	ytd_template := "Give me a short summary on this company's stock performance following year to date stock data titled Stock Performance: .At the end of the summary categorize the company's perfomance into only one of five categories, Very Bad, Bad, Neutral, Good, and Very Good. Data: "
	ytd_inference := getPromptInference(string(queriedInfoAggregate.YTD_Info), ytd_template, "/", "http://localhost:8086")
	promptInference.Stock_Performance = ytd_inference

	// financial health
	fin_template := "Give me a summary on the financial health of this company based on the following financial data titled Financial Health: . The numbers used to represent this data is in scientific notation, so parse it and return the number as a decimal number. At the end of the summary categorize the company's rating into only one of five categories, Very Bad, Bad, Neutral, Good, and Very Good. Data: "
	fin_inference := getPromptInference(string(queriedInfoAggregate.Fin_Info), fin_template, "/", "http://localhost:8086")
	promptInference.Financial_Health = fin_inference

	// news summary
	news_template := "Give me a summary on the news surrounding this company based on the following news data titled News Summary: .At the end of the summary categorize the company's rating into only one of five categories, Very Bad, Bad, Neutral, Good, and Very Good. Data: "
	news_inference := getPromptInference(string(queriedInfoAggregate.News_Info), news_template, "/", "http://localhost:8086")
	promptInference.News_Summary = news_inference

	// company description
	desc_template := "You are an AI who is tasked to summarize financial information for this company ticker. Give me a summary on the company's description. In your summary start the paragraph with the phrase: Company Description: This is the data you will work with: "
	desc_inference := getPromptInference(string(queriedInfoAggregate.Desc_Info), desc_template, "/", "http://localhost:8086")
	promptInference.Company_Desc = desc_inference

	// Return the PromptInference json object as the response
	w.Header().Set("Content-Type", "application/json")
	PromptInference_json, err := json.Marshal(promptInference)
	if err != nil {
		log.Println(err)
	}
	w.Write([]byte(PromptInference_json))

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
