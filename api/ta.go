package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fineas/pkg/serviceauth"
	"fmt"
	"io/ioutil"
	"log"
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
func TechnicalAnalysisService(w http.ResponseWriter, r *http.Request) {

	type TA struct {
		Ticker    string
		StockInfo string
		SMA       string
		MACD      string
		RSI       string
		EMA       string
	}

	type taLOG struct {
		Timestamp       time.Time
		ExecutionTimeMs float32
		RequestIP       string
		EventSequence   []string
	}

	type taOUTPUT struct {
		Result string
	}

	var taLog taLOG
	var eventSequenceArray []string
	var ta TA
	var output taOUTPUT

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
	taLog.RequestIP = ip

	// secure service with pass key hash
	PASS_KEY := os.Getenv("PASS_KEY")
	hash := sha256.New()
	hash.Write([]byte(PASS_KEY))
	getPassHash := hash.Sum(nil)
	passHash := hex.EncodeToString(getPassHash)
	serviceauth.ServiceAuthMiddleware(w, r, eventSequenceArray, passHash)

	MONGO_DB_LOGGER_PASSWORD := os.Getenv("MONGO_DB_LOGGER_PASSWORD")
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	mongoURI := "mongodb+srv://kobenaidun:" + MONGO_DB_LOGGER_PASSWORD + "@cluster0.z9znpv9.mongodb.net/?retryWrites=true&w=majority"
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)
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

	API_KEY := os.Getenv("API_KEY")
	WRITE_KEY := os.Getenv("WRITE_KEY")
	c := polygon.New(API_KEY)

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
	ta.Ticker = ticker
	eventSequenceArray = append(eventSequenceArray, "ticker collected \n")

	if (writeKey == WRITE_KEY) && (len(writeKey) != 0) {
		fmt.Println("write key correct")
		db := client.Database("FinancialInformation")
		db_collection := db.Collection("RawInformation")
		var existingDocument bson.M
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
			eventSequenceArray = append(eventSequenceArray, "found stk info in database \n")
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("400 Bad Request"))
			return
		} else if err == mongo.ErrNoDocuments {
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
			log.Println("Error finding document:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}
	}

	ta.StockInfo, err = getSTKData(ticker, passHash)
	if err != nil {
		log.Printf("Error fetching STK data: %v\n", err)
		eventSequenceArray = append(eventSequenceArray, "could not collect stk info \n")
	} else {
		eventSequenceArray = append(eventSequenceArray, "stk info collected \n")
	}

	getTaMACD, err := sendMACD(c, ticker)
	if err != nil {
		log.Println("Error fetching MACD data:", err)
	}
	getTaSMA, err := sendSMA(c, ticker)
	if err != nil {
		log.Println("Error fetching SMA data:", err)
	}
	getTaEMA, err := sendEMA(c, ticker)
	if err != nil {
		log.Println("Error fetching EMA data:", err)
	}
	getTaRSI, err := getRSI(c, ticker)
	if err != nil {
		log.Println("Error fetching RSI data:", err)
	}

	ta.MACD = formatIndicatorResult("MACD", getTaMACD)
	ta.SMA = formatIndicatorResult("SMA", getTaSMA)
	ta.EMA = formatIndicatorResult("EMA", getTaEMA)
	ta.RSI = formatIndicatorResult("RSI", getTaRSI)
	ta.Ticker = removePrefixSuffix(ticker)

	output.Result = fmt.Sprintf("Stock Info: %s, %s, %s, %s, %s", ta.StockInfo, ta.MACD, ta.SMA, ta.EMA, ta.RSI)
	if err != nil {
		eventSequenceArray = append(eventSequenceArray, "could not collect stk info \n")
		fmt.Println(err)
	}
	taJson, err := json.Marshal(output) // marshal the stk struct into json
	if err != nil {
		eventSequenceArray = append(eventSequenceArray, "Error: could not marshal stk struct into json"+err.Error()+"\n")

	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	taLog.ExecutionTimeMs = float32(elapsedTime.Milliseconds())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprint(string(taJson))))
	taLog.Timestamp = time.Now()
	fmt.Println(string(taJson))

	// insert the log into the database
	eventSequenceArray = append(eventSequenceArray, "successfully served ta data \n")
	taLog.EventSequence = eventSequenceArray
	db := client.Database("MicroserviceLogs")
	db_collection := db.Collection("taServiceLogs")
	_, err = db_collection.InsertOne(context.TODO(), taLog)
	if err != nil {
		log.Fatal(err)
	}

}

// new functions to send the request to the Polygon API for TA information

// sends the request to the Polygon API
func sendSMA(c *polygon.Client, ticker string) (*models.GetSMAResponse, error) {
	params := models.GetSMAParams{ // params for recent date stock price
		Ticker: ticker,
	}.WithWindow(50)

	res, err := c.GetSMA(context.Background(), params) // get recent date stock price
	if err != nil {
		log.Println(err)
		return res, err
	}

	return res, err
}

// sends the request to the Polygon API
func sendEMA(c *polygon.Client, ticker string) (*models.GetEMAResponse, error) {
	params := models.GetEMAParams{
		Ticker: ticker,
	}.WithWindow(50)

	// make request
	res, err := c.GetEMA(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	// do something with the result
	return res, err
}

// rounds the decimal to the specified number of decimal places
func sendMACD(c *polygon.Client, ticker string) (*models.GetMACDResponse, error) {
	// set params
	params := models.GetMACDParams{
		Ticker: ticker,
	}.WithShortWindow(12).
		WithLongWindow(26).
		WithSignalWindow(9).
		WithOrder(models.Desc)

	// make request
	res, err := c.GetMACD(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	// do something with the result
	return res, err
}

func getRSI(c *polygon.Client, ticker string) (*models.GetRSIResponse, error) {
	// set params
	params := models.GetRSIParams{
		Ticker: ticker,
	}.WithWindow(14)

	// make request
	res, err := c.GetRSI(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	// do something with the result
	return res, err
}

func formatIndicatorResult(indicator string, value interface{}) string {
	// Assuming the value is of type *models.GetIndicatorResponse
	// where GetIndicatorResponse represents the response type for each indicator
	switch v := value.(type) {
	case *models.GetMACDResponse:
		if len(v.Results.Values) > 0 {
			return fmt.Sprintf("%s: %s %f %s %f %s %f", indicator, "MACD Value: ", v.Results.Values[0].Value, "Signal: ", v.Results.Values[0].Signal, "Histogram: ", v.Results.Values[0].Histogram)
		}
	case *models.GetSMAResponse:
		if len(v.Results.Values) > 0 {
			return fmt.Sprintf("%s: %s %f", indicator, "SMA Value: ", v.Results.Values[0].Value)
		}
	case *models.GetEMAResponse:
		if len(v.Results.Values) > 0 {
			return fmt.Sprintf("%s: %s %f", indicator, "EMA Value: ", v.Results.Values[0].Value)
		}
	case *models.GetRSIResponse:
		if len(v.Results.Values) > 0 {
			return fmt.Sprintf("%s: %s %f", indicator, "RSI Value: ", v.Results.Values[0].Value)
		}
	}

	return fmt.Sprintf("%s: Data not available", indicator)
}

func getSTKData(ticker string, passHash string) (string, error) {
	stkServiceURL := "http://0.0.0.0:8081/stk" // Replace with actual URL
	req, err := http.NewRequest("GET", stkServiceURL, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("ticker", ticker)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	req.Header.Set("Authorization", "Bearer "+passHash)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK HTTP status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Log the raw response
	fmt.Printf("Raw STK response: %s\n", string(body))
	type STKResponse struct {
		Result string
	}
	var stkResp STKResponse
	err = json.Unmarshal(body, &stkResp)
	if err != nil {
		return "", err
	}

	return stkResp.Result, nil
}
