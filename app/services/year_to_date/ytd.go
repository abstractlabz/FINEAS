package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
)

func main() {
	http.HandleFunc("/ytd", ytdService)
	fmt.Println("Server listening on port 8081...")
	log.Println(http.ListenAndServe(":8081", nil))
}

func ytdService(w http.ResponseWriter, r *http.Request) {

	type YTD struct {
		Ticker                          string
		Recent_Date_Stock_Price         float64
		Year_Before_Recent_Stock_Price  float64
		YTD_Recent_Stock_Percent_Change float64
	}

	var ytd YTD

	err := godotenv.Load("../../../.env") // load the .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	API_KEY := os.Getenv("API_KEY")

	c := polygon.New(API_KEY) // Polygon API connection
	queryParams := r.URL.Query()
	ticker := queryParams.Get("ticker")
	ytd.Ticker = ticker

	//get the relevant date components
	current_year := time.Now().Year()
	current_month := time.Now().Month()
	recent_day := time.Now().Day() - 3

	paramsGDOC := models.GetDailyOpenCloseAggParams{ // params for recent date stock price
		Ticker: ticker,
		Date:   models.Date(time.Date(current_year, current_month, recent_day, 0, 0, 0, 0, time.Local)),
	}.WithAdjusted(true)

	res, err := c.GetDailyOpenCloseAgg(context.Background(), paramsGDOC) // get recent date stock price
	if err != nil {
		log.Println(err)
	}
	ytd.Recent_Date_Stock_Price = res.Open

	paramsGDOC = models.GetDailyOpenCloseAggParams{ // params for year before recent date stock price
		Ticker: ticker,
		Date:   models.Date(time.Date(current_year-1, current_month, recent_day, 0, 0, 0, 0, time.Local)),
	}.WithAdjusted(true)

	res1, err1 := c.GetDailyOpenCloseAgg(context.Background(), paramsGDOC) // get year before recent date stock price
	if err1 != nil {
		log.Println(err1)
	}
	ytd.Year_Before_Recent_Stock_Price = res1.Open

	ytd_info_res := (ytd.Recent_Date_Stock_Price -
		ytd.Year_Before_Recent_Stock_Price) /
		ytd.Year_Before_Recent_Stock_Price // calculate the ytd recent stock percent change
	ytd.YTD_Recent_Stock_Percent_Change = ytd_info_res

	ytd_json, err := json.Marshal(ytd) // marshal the ytd struct into json
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprint(string(ytd_json))))
	fmt.Println("Successfully served ytd data for " + ticker)

}
