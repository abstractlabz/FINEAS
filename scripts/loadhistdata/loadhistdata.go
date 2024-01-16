package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	var filePath string
	var ticker string
	var accessKey string
	type HISTDATA struct {
		Info string
	}

	fmt.Print("Enter the path to your txt file: ")
	fmt.Scanln(&filePath)
	fmt.Print("What company ticker does this data apply to?: ")
	fmt.Scanln(&ticker)
	fmt.Print("Enter your API access key: ")
	fmt.Scanln(&accessKey)

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// Convert content to a string
	text := string(content)
	sentences := splitIntoSentences(text)

	chunkSize := 6
	overlap := 2

	for i := 0; i < len(sentences); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(sentences) {
			end = len(sentences)
		}
		chunk := sentences[i:end]
		chunkText := strings.Join(chunk, " ")

		var histData HISTDATA
		histData.Info = chunkText
		jsonPostData, err := json.Marshal(histData)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			continue
		}

		res := postFinancialData(string(jsonPostData), []string{ticker}, accessKey)
		fmt.Println(res)
	}
}

// splitIntoSentences attempts to split text into sentences more accurately.
func splitIntoSentences(text string) []string {
	var sentences []string
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanWords)

	var currentSentence strings.Builder
	for scanner.Scan() {
		word := scanner.Text()
		currentSentence.WriteString(word + " ")

		if isEndOfSentence(word) {
			sentences = append(sentences, strings.TrimSpace(currentSentence.String()))
			currentSentence.Reset()
		}
	}
	if currentSentence.Len() > 0 {
		sentences = append(sentences, strings.TrimSpace(currentSentence.String()))
	}
	for _, sentence := range sentences {
		fmt.Println(sentence)
	}
	return sentences
}

// isEndOfSentence checks if a word ends with a sentence delimiter.
func isEndOfSentence(word string) bool {
	// Check for standard sentence endings
	if strings.HasSuffix(word, ".") || strings.HasSuffix(word, "!") || strings.HasSuffix(word, "?") {
		// Check for ellipses and common abbreviations
		isEllipsis := strings.HasSuffix(word, "...")
		isAbbreviation := regexp.MustCompile(`(?i)\b(?:[a-z]\.|[dr|mr|mrs|ms|prof|inc|llc|llp|ie|eg]\.?)$`).MatchString(word)
		return !isEllipsis && !isAbbreviation
	}
	return false
}

func postFinancialData(dataValue string, eventSequenceArray []string, passHash string) string {

	url := "http://0.0.0.0:6001/ingestor"
	bearerToken := passHash
	infoData := dataValue

	// Create payload as bytes
	payload := []byte(fmt.Sprintf("info=%s", infoData))

	// Create HTTP client
	client := &http.Client{}

	// Create POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)

	}

	// Set Authorization header
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)

	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)

	}
	defer resp.Body.Close()

	fmt.Println("Response:", string(respBody))
	return string(respBody)
}
