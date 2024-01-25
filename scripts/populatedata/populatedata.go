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

	// Add a command-line flag for manual execution
	manualExecution := flag.Bool("manual", false, "Set to true to execute manually")
	executeKBWRITE := flag.Bool("kbwrite", false, "Set to true to execute writes to knowledge base")
	executeMRWRITE := flag.Bool("mrwrite", false, "Set to true to execute writes for market research")
	flag.Parse()

	if *manualExecution && *executeKBWRITE {
		// Manual execution
		fetchData(stockTickers, KB_WRITE_KEY)
		scheduleFetchData(stockTickers, MR_WRITE_KEY)
		scheduleFetchData(stockTickers, KB_WRITE_KEY)
	} else if *manualExecution && *executeMRWRITE {
		// Scheduled execution
		fetchData(stockTickers, MR_WRITE_KEY)
		scheduleFetchData(stockTickers, KB_WRITE_KEY)
		scheduleFetchData(stockTickers, MR_WRITE_KEY)
	} else if (*manualExecution && *executeKBWRITE && *executeMRWRITE) || (*manualExecution && !*executeKBWRITE && !*executeMRWRITE) {
		// Scheduled execution
		fetchData(stockTickers, MR_WRITE_KEY)
		fetchData(stockTickers, KB_WRITE_KEY)
		scheduleFetchData(stockTickers, KB_WRITE_KEY)
		scheduleFetchData(stockTickers, MR_WRITE_KEY)
	} else {
		// Scheduled execution
		scheduleFetchData(stockTickers, KB_WRITE_KEY)
		scheduleFetchData(stockTickers, MR_WRITE_KEY)
	}
}

func scheduleFetchData(stockTickers []string, KB_WRITE_KEY string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		now := time.Now()
		if now.Hour() == 3 && now.Weekday() >= time.Monday && now.Weekday() <= time.Friday {
			fetchData(stockTickers, KB_WRITE_KEY)
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
