package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fineas/pkg/serviceauth"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// handles the fin request
func FinService(w http.ResponseWriter, r *http.Request) {
	type FINLOG struct {
		Timestamp       time.Time
		ExecutionTimeMs float32
		RequestIP       string
		EventSequence   []string
	}

	type finOUTPUT struct {
		Result string
	}

	var finLog FINLOG
	var eventSequenceArray []string
	var output finOUTPUT

	// Load information structures
	startTime := time.Now()
	queryParams := r.URL.Query()
	err := godotenv.Load("../../.env") // Load the .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "Error parsing IP address: %v", err)
		return
	}
	eventSequenceArray = append(eventSequenceArray, "collected request ip \n")
	finLog.RequestIP = ip

	// Secure service with pass key hash
	PASS_KEY := os.Getenv("PASS_KEY")
	hash := sha256.New()
	hash.Write([]byte(PASS_KEY))
	getPassHash := hash.Sum(nil)
	passHash := hex.EncodeToString(getPassHash)
	serviceauth.ServiceAuthMiddleware(w, r, eventSequenceArray, passHash)

	// Connect to MongoDB
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
	}()

	// Polygon API connection
	API_KEY := os.Getenv("API_KEY")
	c := polygon.New(API_KEY)

	// Ticker input checking
	ticker := queryParams.Get("ticker")
	if len(ticker) == 0 {
		log.Println("Missing required parameter 'ticker' in the query string")
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		w.Write([]byte("Error: Bad Request(400), Missing required parameter 'ticker' in the query string."))
		eventSequenceArray = append(eventSequenceArray, "missing ticker \n")
		return
	}

	// Log ticker
	eventSequenceArray = append(eventSequenceArray, "ticker collected \n")

	// Make request to Polygon API for stock financials
	params := models.ListStockFinancialsParams{}.
		WithTicker(ticker)

	iter := c.VX.ListStockFinancials(context.Background(), params)

	// Accumulate financial values
	var collection string
	for iter.Next() {
		res := fmt.Sprint(iter.Item())
		collection += res
	}

	if iter.Err() != nil {
		eventSequenceArray = append(eventSequenceArray, "could not collect financial statements"+iter.Err().Error()+"\n")
	}

	eventSequenceArray = append(eventSequenceArray, "Collected financial statements \n")

	// Integrate financial statement accumulation here
	targetSubstring := "equity_attributable_to_noncontrolling_interest"
	index := strings.Index(collection, targetSubstring)
	if index != -1 {
		collection = collection[:index]
		collection = deleteNumberBeforeUSD(collection)

		// Accumulate financial values
		accumulatedValues, err := accumulateFinancialValues(collection)
		if err != nil {
			eventSequenceArray = append(eventSequenceArray, "error accumulating financial values: "+err.Error()+"\n")
		} else {
			// Assign the accumulated values to the 'collection' variable
			collection = accumulatedValues
		}

		// Log the accumulated values for debugging
		collection = "The balance sheet for " + ticker + " is " + collection
	} else {
		eventSequenceArray = append(eventSequenceArray, "could not properly collect financial statements \n")
	}

	// Log execution time
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	finLog.ExecutionTimeMs = float32(elapsedTime.Milliseconds())

	// Convert accumulated financial values to JSON format
	collection = strings.ReplaceAll(collection, "\n", ", ")
	output.Result = collection

	// Check if information is already in the database
	db := client.Database("FinancialInformation")
	db_collection := db.Collection("RawInformation")

	// Try to find the document in the database
	var existingDocument bson.M

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
		eventSequenceArray = append(eventSequenceArray, "found fin info in database \n")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("400 Bad Request"))
		return
	} else if err == mongo.ErrNoDocuments {
		// Document not found, insert it into the database
		eventSequenceArray = append(eventSequenceArray, "could not find fin info in database \n")
		_, err := db_collection.InsertOne(context.Background(), bsonData)
		if err != nil {
			eventSequenceArray = append(eventSequenceArray, "could not insert fin info into database \n")
			log.Println("Error inserting document:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}
		eventSequenceArray = append(eventSequenceArray, "successfully inserted fin info into database \n")
	} else {
		// Other error occurred during the FindOne operation
		log.Println("Error finding document:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 Internal Server Error"))
		return
	}

	collection = "The balance sheet for " + ticker + " is " + collection + ", as of date and time " + time.Now().Format("01-02-2006 15:04:05")
	output.Result = collection

	finJson, err := json.Marshal(output) // marshal the stk struct into json
	if err != nil {
		eventSequenceArray = append(eventSequenceArray, "Error: could not marshal stk struct into json"+err.Error()+"\n")

	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(finJson)) // Return the accumulated financial values in JSON format
	finLog.Timestamp = time.Now()

	// Insert the log into the database
	eventSequenceArray = append(eventSequenceArray, "successfully served financials data \n")
	db = client.Database("MicroserviceLogs")
	DBcollection := db.Collection("FinancialsServiceLogs")
	_, err = DBcollection.InsertOne(context.TODO(), finLog)
	if err != nil {
		log.Fatal(err)
	}
}

func accumulateFinancialValues(input string) (string, error) {
	re := regexp.MustCompile(`\b([A-Za-z\s]+)\s+USD\s+([0-9e\+\-\.]+)\s*\}`)
	matches := re.FindAllStringSubmatch(input, -1)

	accumulatedValues := make(map[string]float64)
	for _, match := range matches {
		account := match[1]
		value := match[2]

		decimalValue, err := convertScientificToDecimal(value)
		if err != nil {
			return "", err
		}

		floatValue, err := strconv.ParseFloat(decimalValue, 64)
		if err != nil {
			return "", err
		}

		// Accumulate the values for each account
		accumulatedValues[account] += floatValue
	}

	var result strings.Builder
	first := true
	for account, value := range accumulatedValues {
		if !first {
			result.WriteString(", ")
		} else {
			first = false
		}
		formattedValue := formatCurrency(value)
		result.WriteString(fmt.Sprintf(`"%s": "%s"`, account, formattedValue))
	}

	return result.String(), nil
}

func formatCurrency(amount float64) string {
	formattedValue := fmt.Sprintf("$%.2f", amount)

	parts := strings.Split(formattedValue, ".")
	integralPart := parts[0]

	// Add commas to the integral part
	integralPartWithCommas := addCommasToIntegralPart(integralPart)

	// Combine the integral part with commas and the decimal part
	return integralPartWithCommas + "." + parts[1]
}

func addCommasToIntegralPart(integralPart string) string {
	n := len(integralPart)
	if n <= 3 {
		return integralPart
	}

	var result strings.Builder
	for i, ch := range integralPart {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(ch)
	}

	return result.String()
}

func deleteNumberBeforeUSD(input string) string {
	re := regexp.MustCompile(`(\d+)\s+USD`)
	output := re.ReplaceAllString(input, "USD")
	return output
}

func convertScientificToDecimal(scientificNotation string) (string, error) {
	f, _, err := big.ParseFloat(scientificNotation, 10, 0, big.ToNearestEven)
	if err != nil {
		return "", err
	}
	return f.Text('f', -1), nil
}
