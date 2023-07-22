package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// entry point
func main() {
	http.HandleFunc("/news", newsService)
	fmt.Println("Server started at http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

// handles the news request
func newsService(w http.ResponseWriter, r *http.Request) {
	ticker := r.URL.Query().Get("ticker")
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
