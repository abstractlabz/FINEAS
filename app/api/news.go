package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// handles the news request
func newsService(w http.ResponseWriter, r *http.Request) {

	type NEWSLOG struct {
		ExecutionTimeMs float32
		RequestIP       string
		EventSequence   []string
	}

	var newsLog NEWSLOG
	var eventSequenceArray []string

	//load information structures
	startTime := time.Now()
	queryParams := r.URL.Query()
	err := godotenv.Load("../../.env") // load the .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "Error parsing IP address: %v", err)
		return
	}
	eventSequenceArray = append(eventSequenceArray, "collected request ip \n")
	newsLog.RequestIP = ip

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
		eventSequenceArray = append(eventSequenceArray, "passhash unauthorized \n")
		return
	}
	eventSequenceArray = append(eventSequenceArray, "passhash checked \n")

	// connnect to mongodb
	MONGO_DB_LOGGER_PASSWORD := os.Getenv("MONGO_DB_LOGGER_PASSWORD")
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	mongoURI := "mongodb+srv://kobenaidun:" + MONGO_DB_LOGGER_PASSWORD + "@cluster0.z9znpv9.mongodb.net/?retryWrites=true&w=majority"
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		eventSequenceArray = append(eventSequenceArray, "could not connect to database client \n")
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			eventSequenceArray = append(eventSequenceArray, "could not connect to database \n")
			panic(err)
		}
	}()
	eventSequenceArray = append(eventSequenceArray, "connected to database client \n")

	// Send a ping to confirm a successful connection
	if err := client.Database("MicroserviceLogs").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		eventSequenceArray = append(eventSequenceArray, "could not successfully ping the database \n")
		panic(err)
	}
	eventSequenceArray = append(eventSequenceArray, "successfully pinged database \n")

	// ticker input checking
	ticker := queryParams.Get("ticker")
	if len(ticker) == 0 {
		log.Println("Missing required parameter 'ticker' in the query string")
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		w.Write([]byte("Error: Bad Request(400), Missing required parameter 'ticker' in the query string."))
		eventSequenceArray = append(eventSequenceArray, "missing ticker \n")
		return
	}

	//log ticker
	eventSequenceArray = append(eventSequenceArray, "ticker collected \n")

	entryURL := "https://www.google.com/finance/quote/"
	entryTickerURL := entryURL + ticker
	tickerLinks, err := scrapeTickerLinks(entryTickerURL, ticker)
	if err != nil {
		http.Error(w, "Failed to scrape data", http.StatusInternalServerError)
		eventSequenceArray = append(eventSequenceArray, "Failed to scrape ticker links \n")
		return
	}
	eventSequenceArray = append(eventSequenceArray, "Successfully scraped ticker links \n")

	subLink := tickerLinks[1][1:]

	scrapeTickerURL := "https://www.google.com/finance" + subLink

	textFromDiv, err := scrapeTextFromDiv(scrapeTickerURL)
	if err != nil {
		http.Error(w, "Failed to scrape data", http.StatusInternalServerError)
		eventSequenceArray = append(eventSequenceArray, "Failed to scrape data \n")
		return
	}
	eventSequenceArray = append(eventSequenceArray, "Successfully scraped data \n")

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	newsLog.ExecutionTimeMs = float32(elapsedTime.Milliseconds())
	w.Header().Set("Content-Type", "text/plain")
	fmt.Println(textFromDiv)
	w.Write([]byte(fmt.Sprint(textFromDiv)))
	// insert the log into the database
	eventSequenceArray = append(eventSequenceArray, "successfully served news data for: "+ticker+"\n")
	newsLog.EventSequence = eventSequenceArray
	db := client.Database("MicroserviceLogs")
	DBcollection := db.Collection("NewsServiceLogs")
	_, err = DBcollection.InsertOne(context.TODO(), newsLog)
	if err != nil {
		log.Fatal(err)
	}

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
