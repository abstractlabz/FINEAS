package main

import (
	"context"
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
	http.HandleFunc("/news", newsService)
	fmt.Println("Server listening on port 8083...")
	log.Println(http.ListenAndServe(":8083", nil))
}

func newsService(w http.ResponseWriter, r *http.Request) {

	err := godotenv.Load("../../../.env") // load the .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	API_KEY := os.Getenv("API_KEY")

	c := polygon.New(API_KEY) // Polygon API connection
	queryParams := r.URL.Query()
	ticker := queryParams.Get("ticker")

	//get the relevant date components
	current_year := time.Now().Year()
	current_month := time.Now().Month()
	recent_day := time.Now().Day()

	//
	res := sendRequestWithParamsInfo(c, ticker, current_year, current_month, recent_day)
	for res == "" || res == " " || res == "nil" {
		recent_day = recent_day - 1
		res = sendRequestWithParamsInfo(c, ticker, current_year, current_month, recent_day)
	}

	fmt.Println(res)

	// return the res as the response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprint(res)))

}

func sendRequestWithParamsInfo(c *polygon.Client, ticker string, currentYear int, currentMonth time.Month, recentDay int) string {
	params := models.ListTickerNewsParams{}.
		WithTicker(models.GTE, ticker).
		WithPublishedUTC(models.GTE, models.Millis(time.Date(currentYear, currentMonth,
			recentDay, 0, 0, 0, 0, time.UTC))).
		WithSort(models.PublishedUTC).
		WithOrder(models.Asc).
		WithLimit(1000)

	// make request
	iter := c.ListTickerNews(context.Background(), params)

	// do something with the result
	var collection string
	for iter.Next() {
		res := fmt.Sprint(iter.Item())
		collection += res
	}

	if iter.Err() != nil {
		return "nil"
	}

	return collection
}
