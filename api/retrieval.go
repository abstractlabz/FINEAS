package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostDataInfo struct {
	Ticker            string `bson:"Ticker"`
	StockPerformance  string `bson:"StockPerformance"`
	FinancialHealth   string `bson:"FinancialHealth"`
	NewsSummary       string `bson:"NewsSummary"`
	CompanyDesc       string `bson:"CompanyDesc"`
	TechnicalAnalysis string `bson:"TechnicalAnalysis"`
}

func RetrieveData(w http.ResponseWriter, r *http.Request) {
	// connnect to mongodb
	MONGO_DB_LOGGER_PASSWORD := os.Getenv("MONGO_DB_LOGGER_PASSWORD")
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://kobenaidun:" + MONGO_DB_LOGGER_PASSWORD + "@cluster0.z9znpv9.mongodb.net/?retryWrites=true&w=majority").SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	// Assume ticker is passed as a query parameter
	ticker := r.URL.Query().Get("ticker")
	if ticker == "" {
		http.Error(w, "Ticker not provided", http.StatusBadRequest)
		return
	}

	var postDataInfo PostDataInfo

	// Assuming the collection is named "tickerslist"
	collection := client.Database("FinancialInformation").Collection("TickersList")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	// Find the document
	err = collection.FindOne(ctx, bson.M{"Ticker": ticker}).Decode(&postDataInfo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Ticker not found, return empty JSON
			fmt.Fprint(w, "{}")
			return
		}
		http.Error(w, "Error fetching data from database", http.StatusInternalServerError)
		return
	}

	// Convert the result to JSON and return it
	jsonData, err := json.Marshal(postDataInfo)
	if err != nil {
		http.Error(w, "Error marshalling data to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
