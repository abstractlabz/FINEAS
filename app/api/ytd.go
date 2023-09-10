package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// handles the ytd request
func ytdService(w http.ResponseWriter, r *http.Request) {

	type YTD struct {
		Ticker                      string
		RecentDateStockPrice        float64
		YearBeforeRecentStockPrice  float64
		YTDRecentStockPercentChange float64
	}

	type YTDLOG struct {
		Timestamp       time.Time
		ExecutionTimeMs float32
		RequestIP       string
		EventSequence   []string
	}

	var ytdLog YTDLOG
	var eventSequenceArray []string
	var ytd YTD

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
	ytdLog.RequestIP = ip

	// secure service with pass key hash
	PASS_KEY := os.Getenv("PASS_KEY")
	hash := sha256.New()
	hash.Write([]byte(PASS_KEY))
	getPassHash := hash.Sum(nil)
	passHash := hex.EncodeToString(getPassHash)
	ytdAuthMiddleware(w, r, eventSequenceArray, passHash)

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
	eventSequenceArray = append(eventSequenceArray, "connected to database client \n")

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			eventSequenceArray = append(eventSequenceArray, "could not connect to database \n")
			panic(err)
		}
		eventSequenceArray = append(eventSequenceArray, "connected to database \n")
	}()

	// polygon API connection
	API_KEY := os.Getenv("API_KEY")
	c := polygon.New(API_KEY)

	// ticker input checking
	ticker := queryParams.Get("ticker")
	if len(ticker) == 0 {
		log.Println("Missing required parameter 'ticker' in the query string")
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		w.Write([]byte("Error: Bad Request(400), Missing required parameter 'ticker' in the query string."))
		eventSequenceArray = append(eventSequenceArray, "missing ticker \n")
		return
	}
	fmt.Println(ticker)
	ytd.Ticker = ticker
	eventSequenceArray = append(eventSequenceArray, "ticker collected \n")
	//get the previous closing price
	res, err := sendPreviousCloseInfo(c, ticker)
	ytd.RecentDateStockPrice = res.Results[0].Close

	//calculating dates
	currentYear := time.Now().Year()
	currentMonth := time.Now().Month()
	recentDay := time.Now().Day()

	//gets the year before recent date stock price
	res1, err := sendClosingPriceAtDate(c, ticker, currentYear-1, currentMonth, recentDay)
	//sends request until the stock price is available
	count2 := 0
	for err != nil && res1.Open == 0 {
		if count2 == 7 {
			break
		}
		if recentDay <= 0 {
			recentDay = 31
			currentMonth = currentMonth - 1
			if currentMonth <= 0 {
				currentYear = currentYear - 1
				currentMonth = 12
			}

		}
		res1, err = sendClosingPriceAtDate(c, ticker, currentYear, currentMonth, recentDay)
		recentDay -= 1
		count2 += 1
	}
	ytd.YearBeforeRecentStockPrice = res1.Open
	eventSequenceArray = append(eventSequenceArray, "year before recent date stock price found \n")

	//calculate the ytd recent stock percent change
	ytdInfoRes := roundDecimal(((ytd.RecentDateStockPrice-
		ytd.YearBeforeRecentStockPrice)/
		ytd.YearBeforeRecentStockPrice)*100, 2) // calculate the ytd recent stock percent change
	ytd.YTDRecentStockPercentChange = ytdInfoRes
	eventSequenceArray = append(eventSequenceArray, "ytd recent stock percent change calculated \n")

	ytdJson, err := json.Marshal(ytd) // marshal the ytd struct into json
	if err != nil {
		eventSequenceArray = append(eventSequenceArray, "Error: could not marshal ytd struct into json"+err.Error()+"\n")

	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	ytdLog.ExecutionTimeMs = float32(elapsedTime.Milliseconds())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprint(string(ytdJson))))
	ytdLog.Timestamp = time.Now()
	fmt.Println(string(ytdJson))

	// insert the log into the database
	eventSequenceArray = append(eventSequenceArray, "successfully served ytd data \n")
	ytdLog.EventSequence = eventSequenceArray
	db := client.Database("MicroserviceLogs")
	collection := db.Collection("YtdServiceLogs")
	_, err = collection.InsertOne(context.TODO(), ytdLog)
	if err != nil {
		log.Fatal(err)
	}

}

// sends the request to the Polygon API
func sendClosingPriceAtDate(c *polygon.Client, ticker string, currentYear int, currentMonth time.Month, recentDay int) (*models.GetDailyOpenCloseAggResponse, error) {
	paramsGDOC := models.GetDailyOpenCloseAggParams{ // params for recent date stock price
		Ticker: ticker,
		Date:   models.Date(time.Date(currentYear, currentMonth, recentDay, 0, 0, 0, 0, time.Local)),
	}.WithAdjusted(true)

	res, err := c.GetDailyOpenCloseAgg(context.Background(), paramsGDOC) // get recent date stock price
	if err != nil {
		log.Println(err)
		return res, err
	}

	return res, err
}

// sends the request to the Polygon API
func sendPreviousCloseInfo(c *polygon.Client, ticker string) (*models.GetPreviousCloseAggResponse, error) {
	paramsGDOC := models.GetPreviousCloseAggParams{ // params for recent date stock price
		Ticker: ticker,
	}.WithAdjusted(true)

	res, err := c.GetPreviousCloseAgg(context.Background(), paramsGDOC) // get recent date stock price
	if err != nil {
		log.Println(err)
		return res, err
	}

	return res, err
}

// rounds the decimal to the specified number of decimal places
func roundDecimal(number float64, decimalPlaces int) float64 {
	shift := math.Pow(10, float64(decimalPlaces))
	return math.Round(number*shift) / shift
}

func ytdAuthMiddleware(w http.ResponseWriter, r *http.Request, eventSequenceArray []string, passHash string) bool {

	// Get the Authorization header value
	authHeader := r.Header.Get("Authorization")

	// Check if the Authorization header is present and starts with "Bearer "
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		eventSequenceArray = append(eventSequenceArray, "passhash unauthorized \n")
		return false
	}

	// Extract the token from the Authorization header
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Perform token validation (e.g., check if it's a valid token)
	if token != passHash {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		eventSequenceArray = append(eventSequenceArray, "passhash unauthorized \n")
		return false
	}

	eventSequenceArray = append(eventSequenceArray, "passhash passed \n")
	return true

}