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
	http.HandleFunc("/desc", descriptioService)
	fmt.Println("Server listening on port 8084...")
	log.Println(http.ListenAndServe(":8084", nil))
}

func descriptioService(w http.ResponseWriter, r *http.Request) {

	err := godotenv.Load("../../../.env") // load the .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	API_KEY := os.Getenv("API_KEY")

	c := polygon.New(API_KEY) // Polygon API connection
	queryParams := r.URL.Query()
	ticker := queryParams.Get("ticker")
	fmt.Println(ticker)

	// set params
	params := models.GetTickerDetailsParams{
		Ticker: "AAPL",
	}.WithDate(models.Date(time.Date(2023, 3, 9, 0, 0, 0, 0, time.UTC)))

	// make request
	res, err := c.GetTickerDetails(context.Background(), params)
	if err != nil {
		log.Println(err)
	}

	// do something with the result
	log.Print(res)

}
