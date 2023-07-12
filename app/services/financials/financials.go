package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
)

func main() {
	http.HandleFunc("/fin", finService)
	fmt.Println("Server listening on port 8082...")
	log.Println(http.ListenAndServe(":8082", nil))
}

func finService(w http.ResponseWriter, r *http.Request) {

	err := godotenv.Load("../../../.env") // load the .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	API_KEY := os.Getenv("API_KEY")

	queryParams := r.URL.Query()
	ticker := queryParams.Get("ticker")
	fmt.Println(ticker)

	c := polygon.New(API_KEY) // Polygon API connection

	//
	params := models.ListStockFinancialsParams{}.
		WithTicker(ticker)

	// make request
	iter := c.VX.ListStockFinancials(context.Background(), params)

	// do something with the result

	var collection string
	for iter.Next() {
		res := fmt.Sprint(iter.Item())
		collection += res
	}

	if iter.Err() != nil {
		fmt.Println(iter.Err())
	}
	//
	fmt.Println(collection)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprint(collection)))

}
