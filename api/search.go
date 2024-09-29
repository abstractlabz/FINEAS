package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fineas/pkg/serviceauth"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

// SearchResponse represents the response structure
type SearchResponse struct {
	Results []string `json:"results"`
}

// SearchHandler handles the search request
func SearchHandler(w http.ResponseWriter, r *http.Request) {

	//load information structures
	//startTime := time.Now()
	err := godotenv.Load(".env") // load the .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	eventSequenceArray := []string{}
	eventSequenceArray = append(eventSequenceArray, "collected request ip \n")

	// secure service with pass key hash
	PASS_KEY := os.Getenv("PASS_KEY")
	//WRITE_KEY := os.Getenv("WRITE_KEY")
	hash := sha256.New()
	hash.Write([]byte(PASS_KEY))
	getPassHash := hash.Sum(nil)
	passHash := hex.EncodeToString(getPassHash)
	serviceauth.ServiceAuthMiddleware(w, r, eventSequenceArray, passHash)

	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Get the search query from the form data
	query := r.FormValue("query")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	// Construct the Google search URL
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query))

	// Make a GET request to the Google search URL
	resp, err := http.Get(searchURL)
	if err != nil {
		http.Error(w, "Error fetching search results", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body", http.StatusInternalServerError)
		return
	}

	// Log the raw HTML response (for debugging purposes)
	fmt.Println(string(body))

	// Create an empty array of strings
	results := []string{}

	// Create the response object
	response := SearchResponse{
		Results: results,
	}

	// Marshal the response object to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error marshaling response to JSON", http.StatusInternalServerError)
		return
	}

	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(responseJSON)
}
