package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

type PromptPayload struct {
	Prompt string `json:"prompt"`
}

func ChatbotQuery() gin.HandlerFunc {
	router := gin.Default()
	router.Use(corsMiddleware())

	PASS_KEY := os.Getenv("PASS_KEY")
	OPEN_AI_API_KEY := os.Getenv("OPEN_AI_API_KEY")
	PINECONE_API_KEY := os.Getenv("PINECONE_API_KEY")

	// Initialize Pinecone client
	ClientParams := pinecone.NewClientParams{
		ApiKey: PINECONE_API_KEY,
		Host:   "https://pre-alpha-vectorstore-prd-uajrq2f.svc.aped-4627-b74a.pinecone.io", // optional
	}

	//Initialize Pinecone index
	indexParams := pinecone.NewIndexConnParams{
		Host: "https://pre-alpha-vectorstore-prd-uajrq2f.svc.aped-4627-b74a.pinecone.io",
	}
	// Initialize Pinecone client
	pineconeClient, err := pinecone.NewClient(ClientParams)
	if err != nil {
		log.Fatal("Failed to initialize Pinecone client:", err)
	}
	index, err := pineconeClient.Index(indexParams)
	if err != nil {
		log.Fatal("Failed to initialize Pinecone index:", err)
	}

	router.POST("/chat", func(c *gin.Context) {
		var jsonData PromptPayload
		if err := c.BindJSON(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		passhash := c.GetHeader("Authorization")[7:]
		sha256Hash := sha256.Sum256([]byte(PASS_KEY))
		HASH_KEY := hex.EncodeToString(sha256Hash[:])

		if passhash != HASH_KEY {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		// Embedding and context retrieval
		queryVector, err := embedQuery(jsonData.Prompt, OPEN_AI_API_KEY)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to embed query"})
			return
		}

		context, err := index.QueryByVectorValues(context.Background(), &pinecone.QueryByVectorValuesRequest{
			Vector: queryVector,
			TopK:   7,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query Pinecone index"})
			return
		}

		// Convert context to a string representation
		contextString, err := json.Marshal(context)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert context to string"})
			return
		}

		searchInformation, err := getSearchQuery(jsonData.Prompt, HASH_KEY)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch search information"})
			return
		}

		promptPayload := PromptPayload{
			Prompt: fmt.Sprintf(`You are an AI assistant named Fineas tasked with giving stock market alpha to retail investors
			by summarizing and analyzing financial information in the form of market research. When displaying numbers, show two decimal places.
			Your response will answer the following prompt using structured informative headers, short paragraph segments, annotations, and bullet points for the given financial data. 
			If relevant to the prompt, include general company information such as background history, founder history, current leadership, product history, business segments and their revenue contributions, and anything else pertinent like M&A transactions.
			You will also attach annotation information only defined within the annotations section of this prompt throughout response segments in the text.
			\n\nPROMPT:\n%s\n\n
			The following is the only data context and annotations data from which you will answer this prompt. Only use the annotations which are relevant to the prompt. Ignore the irrelevant annotations and don't include them in your response nor make any reference to them.
			only based off of the most relevant information with 250 words maximum.:
			\n\nANNOTATIONS:\n%s
			\n\nCONTEXT:\n%s`, jsonData.Prompt, searchInformation, contextString),
		}

		LLM_SERVICE_URL := "http://0.0.0.0:8090"
		url := LLM_SERVICE_URL + "/llm"
		headers := map[string]string{
			"Authorization": "Bearer " + HASH_KEY,
			"Content-Type":  "application/json",
		}

		chatResponse, err := fetchChatResponse(url, headers, promptPayload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch response from LLM service"})
			return
		}

		c.String(http.StatusOK, chatResponse)
		c.String(http.StatusOK, chatResponse)
	})

	return nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func embedQuery(prompt, apiKey string) ([]float32, error) {
	url := "https://api.openai.com/v1/embeddings"

	// Create the request payload
	payload := map[string]interface{}{
		"input": prompt,
		"model": "text-embedding-ada-002", // Example model, replace with the appropriate one
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embedding data found")
	}

	return response.Data[0].Embedding, nil
}

func getSearchQuery(rawData, passhash string) (string, error) {
	data := map[string]string{"query": rawData}
	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", "http://localhost:8070/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+passhash)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

func fetchChatResponse(url string, headers map[string]string, payload PromptPayload) (string, error) {
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}
