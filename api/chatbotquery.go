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
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

type PromptPayload struct {
	Prompt string `json:"prompt"`
}

func prettifyStruct(obj interface{}) string {
	bytes, _ := json.MarshalIndent(obj, "", "  ")
	return string(bytes)
}

func min(a, b string) string {
	if a < b {
		return a
	}
	return b
}

func max(a, b string) string {
	if a > b {
		return a
	}
	return b
}

func ChatbotQuery() http.Handler {

	router := gin.Default()
	router.Use(corsMiddleware())

	PASS_KEY := os.Getenv("PASS_KEY")
	PINECONE_API_KEY := os.Getenv("PINECONE_API_KEY")

	// Initialize Pinecone client
	ClientParams := pinecone.NewClientParams{
		ApiKey: PINECONE_API_KEY,
		Host:   "https://main-uajrq2f.svc.aped-4627-b74a.pinecone.io",
	}

	// Initialize Pinecone index
	indexParams := pinecone.NewIndexConnParams{
		Host: "https://main-uajrq2f.svc.aped-4627-b74a.pinecone.io",
	}

	pineconeClient, err := pinecone.NewClient(ClientParams)
	if err != nil {
		log.Println("Failed to initialize Pinecone client:", err)
		http.Error(nil, "Failed to initialize Pinecone client", http.StatusInternalServerError)
	}
	index, err := pineconeClient.Index(indexParams)
	if err != nil {
		log.Println("Failed to initialize Pinecone index:", err)
		http.Error(nil, "Failed to initialize Pinecone index", http.StatusInternalServerError)
	}

	router.POST("/chat", func(c *gin.Context) {

		var jsonData PromptPayload
		if err := c.BindJSON(&jsonData); err != nil {
			log.Println("Error binding JSON:", err)
			c.String(http.StatusBadRequest, "Internal server error, Binding JSON")
			return
		}

		log.Println("Received prompt:", jsonData.Prompt)

		passhash := c.GetHeader("Authorization")[7:]
		sha256Hash := sha256.Sum256([]byte(PASS_KEY))
		HASH_KEY := hex.EncodeToString(sha256Hash[:])

		if passhash != HASH_KEY {
			log.Println("Unauthorized access attempt with passhash:", passhash)
			c.String(http.StatusUnauthorized, "Unauthorized access")
			return
		}

		// Embedding the user's prompt
		queryVector, err := embedQuery(jsonData.Prompt, PINECONE_API_KEY)
		if err != nil {
			log.Println("Failed to embed query:", err)
			c.String(http.StatusInternalServerError, "Internal Server Error, Generating Embeddings")
			return
		}

		log.Println("Query vector:", queryVector)

		LLM_SERVICE_URL := "http://0.0.0.0:8090"
		url := LLM_SERVICE_URL + "/llm"
		headers := map[string]string{
			"Authorization": "Bearer " + HASH_KEY,
			"Content-Type":  "application/json",
		}

		promptPayload := PromptPayload{
			Prompt: jsonData.Prompt,
		}

		promptPayload.Prompt = fmt.Sprintf(`
			For the following prompt, please only give two comma separated dates in YYYY-MM-DD format using no
			spaces which refer to the context of the following prompt. If a date cannot
			 be extrapolated then just use the current date. Do not respond with anything else other than the two dates in the format YYYY-MM-DD,YYYY-MM-DD.
			Prompt: %s
		 `, promptPayload.Prompt)

		chatResponse, err := fetchChatResponse(url, headers, promptPayload)
		if err != nil {
			log.Println("Failed to fetch response from LLM service:", err)
			c.String(http.StatusInternalServerError, "Internal Server Error Fetching Response")
			return
		}

		//this will be a comma separated string of two dates in YYYY-MM-DD format
		dates := strings.Split(chatResponse, ",")
		if len(dates) != 2 {
			log.Printf("Failed to extract dates from LLM response: %v", chatResponse)
			c.String(http.StatusInternalServerError, "Internal Server Error Extracting Dates")
			return
		}

		date1 := dates[0]
		date2 := dates[1]

		// Calculate minimum and maximum dates
		minDate := min(date1, date2)
		maxDate := max(date1, date2)

		minDate = strings.ReplaceAll(minDate, "-", "")
		maxDate = strings.ReplaceAll(maxDate, "-", "")

		minDateInt, err := strconv.Atoi(minDate)
		if err != nil {
			log.Printf("Failed to convert minDate to integer: %v", err)
		}
		log.Println("Min date integer:", minDateInt)
		log.Println("Max date integer plus 1:", minDateInt+1)

		maxDateInt, err := strconv.Atoi(maxDate)
		if err != nil {
			log.Printf("Failed to convert maxDate to integer: %v", err)
		}
		log.Println("Max date integer:", maxDateInt)
		log.Println("Max date integer plus 1:", maxDateInt+1)

		// Define the metadata filter for Pinecone
		metadataMap := map[string]interface{}{
			"current_date": map[string]interface{}{
				"$gte": minDateInt,
				"$lte": maxDateInt,
			},
		}

		log.Println("Metadata map:", metadataMap)

		// Convert the metadataMap to a struct for Pinecone
		metadataFilter, err := structpb.NewStruct(metadataMap)
		if err != nil {
			log.Println("Failed to create metadata map:", err)
			c.String(http.StatusBadRequest, "Internal server error, Metadata Map")
			return
		}

		ctx := context.Background()
		// Perform a similarity search in Pinecone
		searchLimit := uint32(2) // Number of similar documents to retrieve
		searchRes, err := index.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
			Vector:         queryVector,
			TopK:           searchLimit,
			MetadataFilter: metadataFilter,
		})
		if err != nil {
			log.Println("Failed to perform similarity search:", err)
			c.String(http.StatusInternalServerError, "Internal Server Error Similarity Search")
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
				c.String(http.StatusInternalServerError, "Internal Server Error Fetching Vectors")
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
			c.String(http.StatusInternalServerError, "Internal Server Error Search Info")
			return
		}

		log.Println("Search information:", searchInformation)

		currentDate := time.Now().Format("01-02-2006")
		log.Println("Current date:", currentDate)

		promptfield := fmt.Sprintf(`

		You are an AI assistant named Fineas AI tasked with giving stock market
		alpha to retail investors by summarizing and analyzing financial information in the form of market research.
		When displaying numbers, show two decimal places. Your response will answer the following prompt using structured
		informative headers, short paragraph segments, annotations, and bullet points for the given financial data. If relevant
		to the prompt, include general company information such as background history, founder history, current leadership, product history,
		business segments and their revenue contributions, and anything else pertinent like M&A transactions. You will also attach annotation
		url and title information only defined within the annotations section of this prompt throughout response segments in the text. Please include as
		many relevant url link annotations as possible. Please make sure to include the annotation url and title in the response if it is available.
		If it is not available, please omit it from the response and do not reference it at all.

			CURRENT DATE:
			%s
			
			PROMPT:
			%s

			ANNOTATIONS:
			%s

			CONTEXT:
			%s`, currentDate, jsonData.Prompt, searchInformation, contextString)

		promptPayload = PromptPayload{
			Prompt: promptfield,
		}

		log.Println("Prompt payload:", promptPayload)

		chatResponse, err = fetchChatResponse(url, headers, promptPayload)
		if err != nil {
			log.Println("Failed to fetch response from LLM service:", err)
			c.String(http.StatusInternalServerError, "Failed to fetch response from LLM service")
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
	ctx := context.Background()

	// Create a new Pinecone client
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		log.Printf("Failed to create Pinecone client: %v", err)
		return nil, fmt.Errorf("failed to create Pinecone client: %v", err)
	}

	embeddingModel := "multilingual-e5-large"
	queryParameters := pinecone.EmbedParameters{
		InputType: "query",
		Truncate:  "END",
	}

	// Embed the query using Pinecone's inference API
	queryEmbeddingsResponse, err := pc.Inference.Embed(ctx, &pinecone.EmbedRequest{
		Model:      embeddingModel,
		TextInputs: []string{prompt},
		Parameters: queryParameters,
	})
	if err != nil {
		log.Printf("Failed to embed query: %v", err)
		return nil, fmt.Errorf("failed to embed query: %v", err)
	}

	// Assuming the response contains embeddings in a similar structure
	if len(*queryEmbeddingsResponse.Data) == 0 {
		log.Println("No embedding data found in response")
		return nil, fmt.Errorf("no embedding data found")
	}

	// Log and return the embedding data
	log.Println("Embedding data:", (*queryEmbeddingsResponse.Data)[0].Values)
	return *(*queryEmbeddingsResponse.Data)[0].Values, nil
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
