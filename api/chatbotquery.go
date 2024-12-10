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

func prettifyStruct(obj interface{}) string {
	bytes, _ := json.MarshalIndent(obj, "", "  ")
	return string(bytes)
}

func ChatbotQuery() http.Handler {

	router := gin.Default()
	router.Use(corsMiddleware())

	PASS_KEY := os.Getenv("PASS_KEY")
	OPEN_AI_API_KEY := os.Getenv("OPEN_AI_API_KEY")
	PINECONE_API_KEY := os.Getenv("PINECONE_API_KEY")

	// Initialize Pinecone client
	ClientParams := pinecone.NewClientParams{
		ApiKey: PINECONE_API_KEY,
		Host:   "https://pre-alpha-vectorstore-prd-uajrq2f.svc.aped-4627-b74a.pinecone.io",
	}

	// Initialize Pinecone index
	indexParams := pinecone.NewIndexConnParams{
		Host: "https://pre-alpha-vectorstore-prd-uajrq2f.svc.aped-4627-b74a.pinecone.io",
	}

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
			log.Println("Error binding JSON:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		log.Println("Received prompt:", jsonData.Prompt)

		passhash := c.GetHeader("Authorization")[7:]
		sha256Hash := sha256.Sum256([]byte(PASS_KEY))
		HASH_KEY := hex.EncodeToString(sha256Hash[:])

		if passhash != HASH_KEY {
			log.Println("Unauthorized access attempt with passhash:", passhash)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		// Embedding the user's prompt
		queryVector, err := embedQuery(jsonData.Prompt, OPEN_AI_API_KEY)
		if err != nil {
			log.Println("Failed to embed query:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to embed query"})
			return
		}

		log.Println("Query vector:", queryVector)

		ctx := context.Background()
		// Perform a similarity search in Pinecone
		searchLimit := uint32(5) // Number of similar documents to retrieve
		searchRes, err := index.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
			Vector: queryVector,
			TopK:   searchLimit,
		})
		if err != nil {
			log.Println("Failed to perform similarity search:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform similarity search"})
			return
		}

		log.Println("Search results:", prettifyStruct(searchRes))

		// Once we have vector IDs, we can fetch their metadata to get the text
		var contextData []interface{}
		if len(searchRes.Matches) > 0 {
			vectorIds := make([]string, len(searchRes.Matches))
			for i, match := range searchRes.Matches {
				vectorIds[i] = match.Vector.Id
			}
			fetchRes, err := index.FetchVectors(ctx, vectorIds)
			if err != nil {
				log.Println("Failed to fetch vectors:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vector metadata"})
				return
			}

			// Log the fetched vectors to debug
			log.Println("Fetched vectors:", prettifyStruct(fetchRes))

			// Extract the metadata (which should contain your text)
			for _, vector := range fetchRes.Vectors {
				if vector.Metadata != nil {
					log.Println("Vector metadata:", vector.Metadata)
					contextData = append(contextData, vector.Metadata)
				} else {
					log.Println("No metadata found for vector ID:", vector.Id)
				}
			}
		}

		contextString := prettifyStruct(contextData)
		log.Println("Context string:", contextString)

		searchInformation, err := getSearchQuery(jsonData.Prompt, HASH_KEY)
		if err != nil {
			log.Println("Failed to fetch search information:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch search information"})
			return
		}

		log.Println("Search information:", searchInformation)

		promptPayload := PromptPayload{
			Prompt: fmt.Sprintf(`You are an AI assistant named Fineas tasked with giving stock market alpha to retail investors
by summarizing and analyzing financial information in the form of market research. When displaying numbers, show two decimal places.
Your response will answer the following prompt using structured informative headers, short paragraph segments, annotations, and bullet points for the given financial data. 
If relevant to the prompt, include general company information such as background history, founder history, current leadership, product history, business segments and their revenue contributions, and anything else pertinent like M&A transactions.
You will also attach annotation information only defined within the annotations section of this prompt throughout response segments in the text.
			
PROMPT:
%s

The following is the only data context and annotations data from which you will answer this prompt. Only use the annotations which are relevant to the prompt. Ignore the irrelevant annotations and don't include them in your response nor make any reference to them.
only based off of the most relevant information with 250 words maximum.

ANNOTATIONS:
%s

CONTEXT:
%s`, jsonData.Prompt, searchInformation, contextString),
		}

		log.Println("Prompt payload:", promptPayload)

		LLM_SERVICE_URL := "http://0.0.0.0:8090"
		url := LLM_SERVICE_URL + "/llm"
		headers := map[string]string{
			"Authorization": "Bearer " + HASH_KEY,
			"Content-Type":  "application/json",
		}

		chatResponse, err := fetchChatResponse(url, headers, promptPayload)
		if err != nil {
			log.Println("Failed to fetch response from LLM service:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch response from LLM service"})
			return
		}

		log.Println("Chat response:", chatResponse)
		c.String(http.StatusOK, chatResponse)
	})

	return router
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
		"model": "text-embedding-ada-002",
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
	log.Println(response.Data[0].Embedding)
	return response.Data[0].Embedding, nil
}

func getSearchQuery(rawData, passhash string) (string, error) {
	query := rawData

	type SearchQuery struct {
		Query string `json:"query"`
	}

	searchQuery := SearchQuery{Query: query}
	jsonData, err := json.Marshal(searchQuery)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8070/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+passhash)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	log.Println("Search service response:", string(body))
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Error fetching response from LLM service", err
	}

	log.Println(string(body))
	return string(body), nil
}
