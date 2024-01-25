package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	KB_WRITE_KEY := os.Getenv("KB_WRITE_KEY")
	MR_WRITE_KEY := os.Getenv("MR_WRITE_KEY")

	stockTickers := []string{
		//Tech stocks
		"AAPL",
		"META",
		"GOOG",
		"MSFT",
		"AMZN",
		"NVDA",
		"TSLA",
		"INTC",
		//Financial stocks
		"JPM",
		"BAC",
		"C",
		"WFC",
		"GS",
		"MS",
		"PNC",
		"USB",
		"BK",
		"COF",
		//Industrial stocks
		"GE",
		"BA",
		"ACM",
		"VMC",
		"JNJ",
		"HON",
		"DE",
		"CAT",
	}

	// Command-line flags
	manualExecution := flag.Bool("manual", false, "Set to true to execute manually")
	executeKBWRITE := flag.Bool("kbwrite", false, "Set to true to execute writes to knowledge base")
	executeMRWRITE := flag.Bool("mrwrite", false, "Set to true to execute writes for market research")
	flag.Parse()

	// Determining the mode of operation based on flags
	if *manualExecution {
		// Manual execution
		if *executeKBWRITE {
			fetchData(stockTickers, KB_WRITE_KEY)
		}
		if *executeMRWRITE {
			fetchData(stockTickers, MR_WRITE_KEY)
		}
		if !*executeKBWRITE && !*executeMRWRITE {
			log.Println("No action specified for manual execution. Please set either -kbwrite or -mrwrite.")
		}
	} else {
		// Scheduled execution
		scheduleFetchData(stockTickers, KB_WRITE_KEY)
		scheduleFetchData(stockTickers, MR_WRITE_KEY)
	}
}

func scheduleFetchData(stockTickers []string, writeKey string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			if now.Hour() == 3 && now.Weekday() >= time.Monday && now.Weekday() <= time.Friday {
				fetchData(stockTickers, writeKey)
			}
		}
	}
}

func fetchData(stockTickers []string, KB_WRITE_KEY string) {
	client := &http.Client{}
	base_url := "http://0.0.0.0:8080/"

	for _, ticker := range stockTickers {
		url := base_url + "?" + "ticker=" + ticker + "&writekey=" + KB_WRITE_KEY
		fmt.Println("Requesting data for:", ticker)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Error creating request for %s: %v", ticker, err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error fetching data for %s: %v", ticker, err)
			continue
		}
		defer resp.Body.Close()

		// Handle response here as needed
		// For example, you might read and process the response body
	}
}
