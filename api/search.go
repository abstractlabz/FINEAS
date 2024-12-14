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
	"strings"

	"fineas/pkg/serviceauth"

	"github.com/PuerkitoBio/goquery"
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

// SearchHandler handles the search request and scrapes Google search results
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
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	query := requestData.Query
	log.Printf("Received search query: %s", query)

	// Construct the Google search URL
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query))
	log.Printf("Search URL: %s", searchURL)

	// Make a GET request to the Google search URL with a User-Agent header
	client := &http.Client{}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error fetching search results", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Error fetching search results", http.StatusInternalServerError)
		log.Printf("Error: Received status code %d from Google", resp.StatusCode)
		return
	}

	// Parse the HTML using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		http.Error(w, "Error parsing HTML", http.StatusInternalServerError)
		log.Printf("Error parsing HTML: %v", err)
		return
	}

	// Remove <script> and <style> tags
	doc.Find("script, style").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// Prepare to extract data into structured search results
	var searchResults []SearchResult

	// Extract the search result data
	doc.Find("div.kCrYT a").Each(func(i int, s *goquery.Selection) {
		log.Printf("Extracting data for result %d", i+1)

		// Extract the title (from class 'BNeawe vvjwJb AP7Wnd')
		title := s.Find(".vvjwJb").Text()
		log.Printf("Title for result %d: %s", i+1, title)

		// Extract the URL from the 'href' attribute of the <a> tag
		link, exists := s.Attr("href")
		if exists {
			link = parseActualURL(link)
		}
		log.Printf("URL for result %d: %s", i+1, link)

		// Extract the snippet (from class 'BNeawe s3v9rd AP7Wnd')
		snippet := s.Find(".s3v9rd").Text()
		log.Printf("Snippet for result %d: %s", i+1, snippet)

		// Extract the date (from class 'r0bn4c rQMQod')
		date := s.Find(".r0bn4c.rQMQod").Text()
		log.Printf("Date for result %d: %s", i+1, date)

		// Count the number of non-empty data points (Title, URL, Snippet)
		validDataPoints := 0
		if title != "" {
			validDataPoints++
		}
		if link != "" {
			validDataPoints++
		}
		if snippet != "" {
			validDataPoints++
		}

		// Only append results that have at least 2 data points
		if validDataPoints >= 2 {
			searchResults = append(searchResults, SearchResult{
				Title:   title,
				URL:     link,
				Snippet: snippet,
				Date:    date,
			})
		} else {
			log.Printf("Skipping result %d due to insufficient data points", i+1)
		}
	})

	// Log total number of results found
	log.Printf("Total number of results found: %d", len(searchResults))

	// Create the response structure with both results and refined HTML
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

// Helper function to parse and clean the actual URL from Google's search result
func parseActualURL(rawURL string) string {
	// Split to remove everything before 'url=' and extract the actual URL
	if strings.HasPrefix(rawURL, "/url?") {
		parsed := strings.Split(rawURL, "url=")
		if len(parsed) > 1 {
			// Only keep the actual URL and remove any additional parameters
			cleanURL := strings.Split(parsed[1], "&")[0]
			decodedURL, err := url.QueryUnescape(cleanURL)
			if err != nil {
				log.Printf("Error unescaping URL: %v", err)
				return cleanURL
			}
			return decodedURL
		}
	}
	return rawURL
}
