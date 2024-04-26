package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fineas/pkg/serviceauth"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// handles the stk request
func STKService(w http.ResponseWriter, r *http.Request) {

	type STK struct {
		Ticker                      string
		RecentDateStockPrice        float64
		YearBeforeRecentStockPrice  float64
		stkRecentStockPercentChange float64
	}

	type stkLOG struct {
		Timestamp       time.Time
		ExecutionTimeMs float32
		RequestIP       string
		EventSequence   []string
	}

	type stkOUTPUT struct {
		Result string
	}

	var stkLog stkLOG
	var eventSequenceArray []string
	var stk STK
	var output stkOUTPUT

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
	stkLog.RequestIP = ip

	// secure service with pass key hash
	PASS_KEY := os.Getenv("PASS_KEY")
	hash := sha256.New()
	hash.Write([]byte(PASS_KEY))
	getPassHash := hash.Sum(nil)
	passHash := hex.EncodeToString(getPassHash)
	serviceauth.ServiceAuthMiddleware(w, r, eventSequenceArray, passHash)

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
	WRITE_KEY := os.Getenv("WRITE_KEY")
	c := polygon.New(API_KEY)

	// ticker input checking
	ticker := queryParams.Get("ticker")
	writeKey := queryParams.Get("writekey")

	if len(ticker) == 0 {
		log.Println("Missing required parameter 'ticker' in the query string")
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		w.Write([]byte("Error: Bad Request(400), Missing required parameter 'ticker' in the query string."))
		eventSequenceArray = append(eventSequenceArray, "missing ticker \n")
		return
	}
	fmt.Println(ticker)
	stk.Ticker = ticker
	eventSequenceArray = append(eventSequenceArray, "ticker collected \n")
	//get the previous closing price
	res, err := sendPreviousCloseInfo(c, ticker)
	stk.RecentDateStockPrice = res.Results[0].Close

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
	stk.YearBeforeRecentStockPrice = res1.Open
	eventSequenceArray = append(eventSequenceArray, "year before recent date stock price found \n")

	//calculate the stk recent stock percent change
	stkInfoRes := roundDecimal(((stk.RecentDateStockPrice-
		stk.YearBeforeRecentStockPrice)/
		stk.YearBeforeRecentStockPrice)*100, 2) // calculate the stk recent stock percent change
	stk.stkRecentStockPercentChange = stkInfoRes
	eventSequenceArray = append(eventSequenceArray, "stk recent stock percent change calculated \n")

	// construct the output string
	stkOutput := stk.Ticker + " stock previously closed at " + "$" + fmt.Sprint(stk.RecentDateStockPrice) + "." + "The yearly stock percent change for " + ticker + " is " + fmt.Sprint(stk.stkRecentStockPercentChange)
	output.Result = stkOutput

	if (writeKey == WRITE_KEY) && (len(writeKey) != 0) {
		fmt.Println("write key correct")
		// Check if information is already in the database
		db := client.Database("FinancialInformation")
		db_collection := db.Collection("RawInformation")

		// Try to find the document in the database
		var existingDocument bson.M

		// Convert stkJson to BSON format
		bsonData, err := bson.Marshal(output)
		if err != nil {
			eventSequenceArray = append(eventSequenceArray, "could not marshal stkJson to BSON format \n")
			log.Println("Error marshaling document to BSON:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}

		err = db_collection.FindOne(context.Background(), (bsonData)).Decode(&existingDocument)

		if err == nil {
			// Document found in the database
			eventSequenceArray = append(eventSequenceArray, "found stk info in database \n")
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("400 Bad Request"))
			return
		} else if err == mongo.ErrNoDocuments {
			// Document not found, insert it into the database
			eventSequenceArray = append(eventSequenceArray, "could not find stk info in database \n")
			_, err := db_collection.InsertOne(context.Background(), bsonData)
			if err != nil {
				eventSequenceArray = append(eventSequenceArray, "could not insert stk info into database \n")
				log.Println("Error inserting document:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500 Internal Server Error"))
				return
			}
			eventSequenceArray = append(eventSequenceArray, "successfully inserted stk info into database \n")
		} else {
			// Other error occurred during the FindOne operation
			log.Println("Error finding document:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}
		return
	}

	// update stk output string with date
	stk.Ticker, _ = removePrefixSuffix(stk.Ticker)
	stkOutput = stk.Ticker + " stock previously closed at " + "$" + fmt.Sprint(stk.RecentDateStockPrice) + "." + "The yearly stock percent change for " + ticker + " is " + fmt.Sprint(stk.stkRecentStockPercentChange) + "%" + ", as of date and time " + time.Now().Format("01-02-2006 15:04:05")
	output.Result = stkOutput

	stkJson, err := json.Marshal(output) // marshal the stk struct into json
	if err != nil {
		eventSequenceArray = append(eventSequenceArray, "Error: could not marshal stk struct into json"+err.Error()+"\n")

	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	stkLog.ExecutionTimeMs = float32(elapsedTime.Milliseconds())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprint(string(stkJson))))
	stkLog.Timestamp = time.Now()
	fmt.Println(string(stkJson))

	// insert the log into the database
	eventSequenceArray = append(eventSequenceArray, "successfully served stk data \n")
	stkLog.EventSequence = eventSequenceArray
	db := client.Database("MicroserviceLogs")
	db_collection := db.Collection("stkServiceLogs")
	_, err = db_collection.InsertOne(context.TODO(), stkLog)
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
	params := models.GetPreviousCloseAggParams{
		Ticker: ticker,
	}.WithAdjusted(true)

	// make request
	res, err := c.GetPreviousCloseAgg(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	// do something with the result
	return res, err
}

// rounds the decimal to the specified number of decimal places
func roundDecimal(number float64, decimalPlaces int) float64 {
	shift := math.Pow(10, float64(decimalPlaces))
	return math.Round(number*shift) / shift
}
