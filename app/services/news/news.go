package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	http.HandleFunc("/news", newsService)
	fmt.Println("Server started at http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

func newsService(w http.ResponseWriter, r *http.Request) {
	ticker := r.URL.Query().Get("ticker")
	if ticker == "" {
		http.Error(w, "Please provide a 'ticker' query parameter", http.StatusBadRequest)
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

	fmt.Fprintln(w, textFromDiv)

}

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
