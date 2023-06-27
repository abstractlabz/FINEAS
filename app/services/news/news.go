package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	polygon "github.com/polygon-io/client-go/rest"
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

	fmt.Println(ticker)
	fmt.Println(c)

	// return the res as the response
	w.Header().Set("Content-Type", "application/json")

}
