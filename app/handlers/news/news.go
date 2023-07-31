package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

// entry point
func main() {
	http.HandleFunc("/news", newsService)
	fmt.Println("Server started at http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

// handles the news request
func newsService(w http.ResponseWriter, r *http.Request) {

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

	// ticker input checking
	ticker := queryParams.Get("ticker")
	if len(ticker) == 0 {
		log.Println("Missing required parameter 'ticker' in the query string")
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		w.Write([]byte("Error: Bad Request(400), Missing required parameter 'ticker' in the query string."))
		return
	}

	entryURL := "https://www.google.com/finance/quote/"
	entryTickerURL := entryURL + ticker
	tickerLinks, err := scrapeTickerLinks(entryTickerURL, ticker)
	if err != nil {
		http.Error(w, "Failed to scrape data", http.StatusInternalServerError)
		return
	}

	subLink := tickerLinks[1][1:]

	scrapeTickerURL := "https://www.google.com/finance" + subLink

	textFromDiv, err := scrapeTextFromDiv(scrapeTickerURL)
	if err != nil {
		http.Error(w, "Failed to scrape data", http.StatusInternalServerError)
		return
	}

	fmt.Println(textFromDiv)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Println(textFromDiv)
	w.Write([]byte(fmt.Sprint(textFromDiv)))
	fmt.Println("Successfully served news data for " + ticker)

}

// scrapeTickerLinks scrapes the ticker links from the Google Finance page
func scrapeTickerLinks(url string, ticker string) ([]string, error) {
	var tickerLinks []string

	// Make a GET request to the URL
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Parse the HTML response body
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// Find all anchor tags and extract the links containing the ticker
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if strings.Contains(href, ticker) {
			tickerLinks = append(tickerLinks, href)
		}
	})

	return tickerLinks, nil
}

// scrapeTextFromDiv scrapes the text from Google's top news section
func scrapeTextFromDiv(url string) (string, error) {
	// Make a GET request to the URL
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Parse the HTML response body
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	// Find the text within the div with class "F2KAFc"
	text := ""
	doc.Find("div.F2KAFc").Each(func(i int, s *goquery.Selection) {
		text += s.Text() + "\n"
	})

	return text, nil
}
