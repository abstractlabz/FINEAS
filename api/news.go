package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fineas/pkg/serviceauth"
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
func NewsService(w http.ResponseWriter, r *http.Request) {

	type NEWSLOG struct {
		Timestamp       time.Time
		ExecutionTimeMs float32
		RequestIP       string
		EventSequence   []string
	}

	type newsOUTPUT struct {
		Result string
	}

	var newsLog NEWSLOG
	var output newsOUTPUT
	var eventSequenceArray []string
	var tag string

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
	WRITE_KEY := os.Getenv("WRITE_KEY")
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
	writeKey := queryParams.Get("writekey")

	ticker, tag = removePrefixSuffix(ticker)

	if len(ticker) == 0 {
		log.Println("Missing required parameter 'ticker' in the query string")
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		w.Write([]byte("Error: Bad Request(400), Missing required parameter 'ticker' in the query string."))
		eventSequenceArray = append(eventSequenceArray, "missing ticker \n")
		return
	}

	//log ticker
	eventSequenceArray = append(eventSequenceArray, "ticker collected \n")

	scrapeTickerURL := "https://news.google.com/search?q=" + strings.ToUpper(ticker) + "-news" + tag + "&hl=en-US&gl=US&ceid=US%3Aen"

	textFromDiv, err := scrapeTextFromDiv(scrapeTickerURL, 5)
	if err != nil {
		http.Error(w, "Failed to scrape data", http.StatusInternalServerError)
		eventSequenceArray = append(eventSequenceArray, "Failed to scrape data \n")
		log.Println(ticker, " Failed to scrape data:", err)
		fmt.Println(ticker, " Failed to scrape data:", err)
		w.Write([]byte(ticker + " Failed to scrape news data"))
		w.Write([]byte("500 Internal Server Error"))
		return
	} else {
		eventSequenceArray = append(eventSequenceArray, "Successfully scraped data \n")
	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	newsLog.ExecutionTimeMs = float32(elapsedTime.Milliseconds())

	// Convert accumulated financial values to JSON format
	textFromDiv = strings.ReplaceAll(textFromDiv, "\n", ". ")
	output.Result = textFromDiv

	newsJson, err := json.Marshal(output) // marshal the stk struct into json
	if err != nil {
		eventSequenceArray = append(eventSequenceArray, "Error: could not marshal stk struct into json"+err.Error()+"\n")

	}

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

		err = db_collection.FindOne(context.Background(), bsonData).Decode(&existingDocument)

		if err == nil {
			// Document found in the database
			eventSequenceArray = append(eventSequenceArray, "found news info in database \n")
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("400 Bad Request"))
			return
		} else if err == mongo.ErrNoDocuments {
			// Document not found, insert it into the database
			eventSequenceArray = append(eventSequenceArray, "could not find news info in database \n")
			_, err := db_collection.InsertOne(context.Background(), bsonData)
			if err != nil {
				eventSequenceArray = append(eventSequenceArray, "could not insert news info into database \n")
				log.Println("Error inserting document:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500 Internal Server Error"))
				return
			}
			eventSequenceArray = append(eventSequenceArray, "successfully inserted news info into database \n")
		} else {
			// Other error occurred during the FindOne operation
			log.Println("Error finding document:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(newsJson))
	fmt.Println(string(newsJson))
	newsLog.Timestamp = time.Now()

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

// scrapeTextFromDiv scrapes the text from Google's top news section
func scrapeTextFromDiv(url string, collectionSize int) (string, error) {
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

	text := ""
	doc.Find("a.JtKRv").Each(func(i int, s *goquery.Selection) {
		if i <= collectionSize {
			text += s.Text() + "\n"
		}
	})

	return text, nil
}

func removePrefixSuffix(ticker string) (string, string) {
	// Remove prefix 'x:' if it exists
	if strings.HasPrefix(ticker, "X:") {
		ticker = strings.TrimPrefix(ticker, "X:")

		ticker = strings.TrimSuffix(ticker, "USD")

		return ticker, "-crypto"
	} else {
		return ticker, "-stock"
	}
}
