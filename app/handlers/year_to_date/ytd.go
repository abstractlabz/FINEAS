package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
)

// entry point
func main() {
	http.HandleFunc("/ytd", ytdService)
	fmt.Println("Server listening on port 8081...")
	log.Println(http.ListenAndServe(":8081", nil))
}

// handles the ytd request
func ytdService(w http.ResponseWriter, r *http.Request) {

	type YTD struct {
		Ticker                      string
		RecentDateStockPrice        float64
		YearBeforeRecentStockPrice  float64
		YTDRecentStockPercentChange float64
	}

	var ytd YTD

	//load information structures
	queryParams := r.URL.Query()
	err := godotenv.Load("../../../.env") // load the .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// secure service with pass key hash
	PASS_KEY := os.Getenv("PASS_KEY")
	hash := sha256.New()
	hash.Write([]byte(PASS_KEY))
	getPassHash := hash.Sum(nil)
	passHash := hex.EncodeToString(getPassHash)
	passHashFromRequest := queryParams.Get("passhash")
	if passHash != passHashFromRequest {
		log.Println("Incorrect passhash: unathorized request")
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		w.Write([]byte("Error: Unauthorized(401), Incorrect passhash."))
		return
	}

	// polygon API connection
	API_KEY := os.Getenv("API_KEY")
	c := polygon.New(API_KEY)

	// ticker input checking
	ticker := queryParams.Get("ticker")
	if len(ticker) == 0 {
		log.Println("Missing required parameter 'ticker' in the query string")
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		w.Write([]byte("Error: Bad Request(400), Missing required parameter 'ticker' in the query string."))
		return
	}
	fmt.Println(ticker)
	ytd.Ticker = ticker

	//get the relevant date components
	currentYear := time.Now().Year()
	currentMonth := time.Now().Month()
	recentDay := time.Now().Day() - 1

	//get the recent date stock price
	res := sendRequestWithParamsInfo(c, ticker, currentYear, currentMonth, recentDay)
	//sends request until the stock price is available
	for res.Open == 0 {
		recentDay = recentDay - 1
		res = sendRequestWithParamsInfo(c, ticker, currentYear, currentMonth, recentDay)
	}
	ytd.RecentDateStockPrice = res.Open

	//gets the year before recent date stock price
	res1 := sendRequestWithParamsInfo(c, ticker, currentYear-1, currentMonth, recentDay)
	//sends request until the stock price is available
	for res1.Open == 0 {
		currentYear = currentYear - 1
		res1 = sendRequestWithParamsInfo(c, ticker, currentYear, currentMonth, recentDay)
	}
	ytd.YearBeforeRecentStockPrice = res1.Open

	//calculate the ytd recent stock percent change
	ytdInfoRes := roundDecimal(((ytd.RecentDateStockPrice-
		ytd.YearBeforeRecentStockPrice)/
		ytd.YearBeforeRecentStockPrice)*100, 2) // calculate the ytd recent stock percent change
	ytd.YTDRecentStockPercentChange = ytdInfoRes

	ytdJson, err := json.Marshal(ytd) // marshal the ytd struct into json
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprint(string(ytdJson))))
	fmt.Println(string(ytdJson))
	fmt.Println("Successfully served ytd data for " + ticker)

}

// sends the request to the Polygon API
func sendRequestWithParamsInfo(c *polygon.Client, ticker string, currentYear int, currentMonth time.Month, recentDay int) *models.GetDailyOpenCloseAggResponse {
	paramsGDOC := models.GetDailyOpenCloseAggParams{ // params for recent date stock price
		Ticker: ticker,
		Date:   models.Date(time.Date(currentYear, currentMonth, recentDay, 0, 0, 0, 0, time.Local)),
	}.WithAdjusted(true)

	res, err := c.GetDailyOpenCloseAgg(context.Background(), paramsGDOC) // get recent date stock price
	if err != nil {
		log.Println(err)
	}

	return res
}

// rounds the decimal to the specified number of decimal places
func roundDecimal(number float64, decimalPlaces int) float64 {
	shift := math.Pow(10, float64(decimalPlaces))
	return math.Round(number*shift) / shift
}
