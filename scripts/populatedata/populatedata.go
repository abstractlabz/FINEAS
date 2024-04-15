package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Ticker struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	MR_WRITE_KEY := os.Getenv("MR_WRITE_KEY")
	KB_WRITE_KEY := os.Getenv("KB_WRITE_KEY")

	// Load tickers from JSON file
	var tickers []Ticker
	file, err := os.ReadFile("../../utils/data/tickerslist.json")
	if err != nil {
		log.Fatalf("Error reading tickers list: %v", err)
	}
	err = json.Unmarshal(file, &tickers)
	if err != nil {
		log.Fatalf("Error unmarshaling tickers: %v", err)
	}

	// Convert tickers to a string slice
	var stockTickers []string
	for _, ticker := range tickers {
		stockTickers = append(stockTickers, ticker.Value)
	}

	// Command-line flags
	manualExecution := flag.Bool("manual", false, "Set to true to execute manually")
	executeMRWRITE := flag.Bool("mrwrite", false, "Set to true to execute writes for market research")
	executeKBWRITE := flag.Bool("kbwrite", false, "Set to true to execute writes to knowledge base")
	batchSize := flag.Int("batchsize", 10, "Number of tickers per batch")
	flag.Parse()

	if *manualExecution {
		if *executeMRWRITE {
			sendBatches(stockTickers, MR_WRITE_KEY, *batchSize)
		}
		if *executeKBWRITE {
			sendBatches(stockTickers, KB_WRITE_KEY, *batchSize)
		}
		if !*executeKBWRITE && !*executeMRWRITE {
			log.Println("No action specified for manual execution. Please set either -kbwrite or -mrwrite.")
		}
	} else {
		log.Println("Automated execution is not supported in this script.")
	}
}

func sendBatches(stockTickers []string, writeKey string, batchSize int) {
	batches := createBatches(stockTickers, batchSize)
	for _, batch := range batches {
		fetchData(batch, writeKey)
	}
}

func fetchData(stockTickers []string, writeKey string) {
	client := &http.Client{}
	base_url := "http://0.0.0.0:8080/"
	var wg sync.WaitGroup

	for _, ticker := range stockTickers {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			t = strings.ToUpper(t)
			url := fmt.Sprintf("%s?ticker=%s&writekey=%s", base_url, t, writeKey)
			fmt.Println("Requesting data for:", t)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Printf("Error creating request for %s: %v", t, err)
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error fetching data for %s: %v", t, err)
				return
			}
			// Ensure the response body is closed to free up resources
			resp.Body.Close()
		}(ticker)
	}
	wg.Wait() // Wait for all requests in the batch to complete
}

func createBatches(tickers []string, size int) [][]string {
	var batches [][]string
	for len(tickers) > size {
		tickers, batches = tickers[size:], append(batches, tickers[0:size:size])
	}
	batches = append(batches, tickers)
	return batches
}
