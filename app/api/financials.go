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

	"github.com/joho/godotenv"
	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// handles the fin request
func finService(w http.ResponseWriter, r *http.Request) {

	type FINLOG struct {
		Timestamp       time.Time
		ExecutionTimeMs float32
		RequestIP       string
		EventSequence   []string
	}

	var finLog FINLOG
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
	finLog.RequestIP = ip

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
	eventSequenceArray = append(eventSequenceArray, "connected to database client \n")

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			eventSequenceArray = append(eventSequenceArray, "could not connect to database \n")
			panic(err)
		}
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

	// log ticker
	eventSequenceArray = append(eventSequenceArray, "ticker collected \n")

	params := models.ListStockFinancialsParams{}.
		WithTicker(ticker)

	// make request
	iter := c.VX.ListStockFinancials(context.Background(), params)

	// do something with the result

	var collection string
	for iter.Next() {
		res := fmt.Sprint(iter.Item())
		collection += res
	}

	if iter.Err() != nil {
		eventSequenceArray = append(eventSequenceArray, "could not collect financial statements"+iter.Err().Error()+"\n")

	}
	eventSequenceArray = append(eventSequenceArray, "Collected financial statements \n")

	targetSubstring := "equity_attributable_to_noncontrolling_interest"
	index := strings.Index(collection, targetSubstring)
	if index != -1 {
		collection = collection[:index]
		print(collection)
	} else {
		eventSequenceArray = append(eventSequenceArray, "could not properly collect financial statements \n")
	}

	// log execution time
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	finLog.ExecutionTimeMs = float32(elapsedTime.Milliseconds())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprint(collection)))
	finLog.Timestamp = time.Now()

	// insert the log into the database
	eventSequenceArray = append(eventSequenceArray, "successfully served financials data \n")
	finLog.EventSequence = eventSequenceArray
	db := client.Database("MicroserviceLogs")
	DBcollection := db.Collection("FinancialsServiceLogs")
	_, err = DBcollection.InsertOne(context.TODO(), finLog)
	if err != nil {
		log.Fatal(err)
	}

}
