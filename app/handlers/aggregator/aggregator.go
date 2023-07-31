package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// entry point
func main() {
	http.HandleFunc("/", handleQuoteRequest)
	fmt.Println("Server listening on port 8080...")
	log.Println(http.ListenAndServe(":8080", nil))
}

// handles the ticker request
func handleQuoteRequest(w http.ResponseWriter, r *http.Request) {

	//to represent the aggregate of all financial information
	type QueriedInfoAggregate struct {
		Ticker   string
		YtdInfo  string
		FinInfo  string
		NewsInfo string
		DescInfo string
	}

	//to represent the aggregate of all prompt inferences
	type PromptInference struct {
		StockPerformance string
		FinancialHealth  string
		NewsSummary      string
		CompanyDesc      string
	}

	var promptInference PromptInference

	// Create a new instance of QueriedInfoAggregate
	var queriedInfoAggregate QueriedInfoAggregate

	// Process query string parameters from the request URL
	queryParams := r.URL.Query()
	ticker := queryParams.Get("ticker")
	if len(ticker) == 0 {
		log.Println("Missing required parameter 'ticker' in the query string")
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		w.Write([]byte("Error: Bad Request(400), Missing required parameter 'ticker' in the query string."))
		return
	}
	fmt.Println(ticker)

	// Get the financial information from the services
	// and return it as the response

	err := godotenv.Load("../../../.env") // load the .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	YTD_SERVICE_URL := os.Getenv("YTD_SERVICE_URL")
	FIN_SERVICE_URL := os.Getenv("FIN_SERVICE_URL")
	NEWS_SERVICE_URL := os.Getenv("NEWS_SERVICE_URL")
	DESC_SERVICE_URL := os.Getenv("DESC_SERVICE_URL")
	YTD_TEMPLATE := os.Getenv("YTD_TEMPLATE")
	FIN_TEMPLATE := os.Getenv("FIN_TEMPLATE")
	PASS_KEY := os.Getenv("PASS_KEY")

	// create a 256 sha hash of pass key from env file
	hash := sha256.New()
	hash.Write([]byte(PASS_KEY))
	getPassHash := hash.Sum(nil)
	passHash := hex.EncodeToString(getPassHash)

	fmt.Print("passHash: " + passHash + "\n")

	queriedInfoAggregate.Ticker = ticker
	ytd_info := getFinancialInfo(ticker, "/ytd", YTD_SERVICE_URL, passHash)
	queriedInfoAggregate.YtdInfo = ytd_info

	fin_info := getFinancialInfo(ticker, "/fin", FIN_SERVICE_URL, passHash)
	queriedInfoAggregate.FinInfo = fin_info

	news_info := getFinancialInfo(ticker, "/news", NEWS_SERVICE_URL, passHash)
	queriedInfoAggregate.NewsInfo = news_info

	desc_info := getFinancialInfo(ticker, "/desc", DESC_SERVICE_URL, passHash)
	queriedInfoAggregate.DescInfo = desc_info

	// stock perfomance
	ytdTemplate := YTD_TEMPLATE
	ytdInference := getPromptInference(string(queriedInfoAggregate.YtdInfo), ytdTemplate, "/", "http://localhost:8086", passHash)
	promptInference.StockPerformance = ytdInference

	// financial health
	finTemplate := FIN_TEMPLATE
	finInference := getPromptInference(string(queriedInfoAggregate.FinInfo), finTemplate, "/", "http://localhost:8086", passHash)
	promptInference.FinancialHealth = finInference

	// news summary
	promptInference.NewsSummary = queriedInfoAggregate.NewsInfo

	// company description
	promptInference.CompanyDesc = queriedInfoAggregate.DescInfo

	// Return the PromptInference json object as the response
	w.Header().Set("Content-Type", "application/json")
	promptInferenceJson, err := json.Marshal(promptInference)
	if err != nil {
		log.Println(err)
	}
	w.Write([]byte(promptInferenceJson))

}

// gets the financial information from the polygon.io services
func getFinancialInfo(ticker string, handlerID string, handlerURL string, passHash string) string {

	base_url := handlerURL + handlerID

	// Construct the URL with query parameters
	url := base_url + "?" + "ticker=" + ticker + "&" + "passhash=" + passHash

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

// gets the prompt inference from the LLM service
func getPromptInference(prompt string, template string, handlerID string, handlerURL string, passHash string) string {

	baseUrl := handlerURL + handlerID

	url := baseUrl + "?" + "prompt=" + urlConverter(template+prompt) + "&" + "passhash=" + passHash

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

// converts prompt to a URL compatible format
func urlConverter(_url string) string {

	// Construct the URL with query parameters
	encodedPrompt := url.QueryEscape(_url)
	encodedPrompt = strings.ReplaceAll(encodedPrompt, "+", "%20")
	encodedPrompt = strings.ReplaceAll(encodedPrompt, "%2F", "/")
	encodedPrompt = strings.ReplaceAll(encodedPrompt, "%3A", ":")

	return encodedPrompt
}
