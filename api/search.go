package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"fineas/pkg/serviceauth"

	"github.com/joho/godotenv"
)

// SearchResult defines the structure for a single search result
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Date    string `json:"date"`
}

// Response structure to include both the refined HTML and structured JSON
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// GoogleAPIResponse represents the structure of the Google Custom Search API response
type GoogleAPIResponse struct {
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
}

// SearchHandler handles the search request and fetches Google search results using the Custom Search API
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Load environment variables
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Authenticate request
	eventSequenceArray := []string{}
	eventSequenceArray = append(eventSequenceArray, "collected request ip \n")

	PASS_KEY := os.Getenv("PASS_KEY")
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

	// Parse the JSON body to get the query
	var requestData struct {
		Query string `json:"query"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.Query == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	query := requestData.Query
	log.Printf("Received search query: %s", query)

	// Set up Google Custom Search API parameters
	apiKey := os.Getenv("GOOGLE_API_KEY") // Load from environment variables
	cx := os.Getenv("GOOGLE_CSE_ID")      // Load Custom Search Engine ID from environment variables
	numResults := 10                      // Number of results to fetch
	apiURL := "https://customsearch.googleapis.com/customsearch/v1"

	// Build the request URL
	reqURL, err := url.Parse(apiURL)
	if err != nil {
		http.Error(w, "Error parsing API URL", http.StatusInternalServerError)
		log.Fatalf("Error parsing API URL: %v", err)
	}

	// Add query parameters
	queryParams := reqURL.Query()
	queryParams.Set("key", apiKey)
	queryParams.Set("cx", cx)
	queryParams.Set("q", query)
	queryParams.Set("num", fmt.Sprintf("%d", numResults))
	reqURL.RawQuery = queryParams.Encode()

	// Make the API request
	resp, err := http.Get(reqURL.String())
	if err != nil {
		http.Error(w, "Error making API request", http.StatusInternalServerError)
		log.Fatalf("Error making API request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Error fetching search results", http.StatusInternalServerError)
		log.Printf("Error: Received status code %d from Google API", resp.StatusCode)
		return
	}

	// Parse the API response
	var apiResponse GoogleAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		http.Error(w, "Error parsing API response", http.StatusInternalServerError)
		log.Printf("Error parsing API response: %v", err)
		return
	}

	// Extract search results
	var searchResults []SearchResult
	for _, item := range apiResponse.Items {
		searchResults = append(searchResults, SearchResult{
			Title:   item.Title,
			URL:     item.Link,
			Snippet: item.Snippet,
			Date:    "", // Date field is not provided by the API
		})
	}

	// Log total number of results found
	log.Printf("Total results found: %d", len(searchResults))

	// Create the response structure with the results
	response := SearchResponse{
		Results: searchResults,
	}

	// Convert the response to JSON
	resultsJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error converting results to JSON", http.StatusInternalServerError)
		log.Printf("Error converting results to JSON: %v", err)
		return
	}

	// Set the content type to JSON and return the results
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultsJSON)
}
