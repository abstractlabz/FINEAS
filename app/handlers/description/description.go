package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
)

// entry point
func main() {
	http.HandleFunc("/desc", descriptionService)
	fmt.Println("Server listening on port 8084...")
	log.Println(http.ListenAndServe(":8084", nil))
}

// handles the desc request
func descriptionService(w http.ResponseWriter, r *http.Request) {

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

	//log ticker
	fmt.Println(ticker)

	// set params
	params := models.GetTickerDetailsParams{
		Ticker: ticker,
	}.WithDate(models.Date(time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)))

	// make request
	res, err := c.GetTickerDetails(context.Background(), params)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(res)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprint(res)))

}
