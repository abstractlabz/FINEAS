package api

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strings"

	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

func LLMHandler(router *gin.Engine) {
	router.Use(corsMiddleware())

	// Define the /llm endpoint
	router.POST("/llm", func(c *gin.Context) {
		// Load environment variables
		CLAUDE_API_URL := "https://api.anthropic.com/v1/messages"
		CLAUDE_API_KEY := os.Getenv("CLAUDE_API_KEY")
		PASS_KEY := os.Getenv("PASS_KEY")

		if CLAUDE_API_URL == "" || CLAUDE_API_KEY == "" || PASS_KEY == "" {
			log.Fatal("CLAUDE_API_URL, CLAUDE_API_KEY, and PASS_KEY environment variables are required")
		}

		log.Println("Using CLAUDE_API_KEY:", CLAUDE_API_KEY) // Print the API key for debugging purposes

		// Enable CORS middleware
		router.Use(corsMiddleware())
		// Extract prompt from JSON payload
		var jsonData struct {
			Prompt string `json:"prompt"`
		}
		if err := c.BindJSON(&jsonData); err != nil {
			log.Println("Error binding JSON:", err)
			c.String(http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		log.Println("Received prompt:", jsonData.Prompt)

		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Println("Unauthorized access attempt with authHeader:", authHeader)
			c.String(http.StatusUnauthorized, "Unauthorized access")
			return
		}
		passhash := authHeader[7:]

		// Compute SHA256 hash of PASS_KEY
		sha256Hash := sha256.Sum256([]byte(PASS_KEY))
		HASH_KEY := hex.EncodeToString(sha256Hash[:])

		// Verify the hashed passkey
		if passhash != HASH_KEY {
			log.Println("Unauthorized access attempt with passhash:", passhash)
			c.String(http.StatusUnauthorized, "Unauthorized access")
			return
		}

		if jsonData.Prompt == "" {
			log.Println("Missing prompt parameter")
			c.String(http.StatusBadRequest, "Missing prompt parameter")
			return
		}

		// Log the prompt (optional)
		log.Println("Prompt:", jsonData.Prompt)

		// Prepare the request payload for Claude API
		requestPayload := map[string]interface{}{
			"model":      "claude-3-5-sonnet-20241022",
			"max_tokens": 4096,
			"messages": []map[string]string{
				{"role": "user", "content": jsonData.Prompt},
			},
		}

		payloadBytes, err := json.Marshal(requestPayload)
		if err != nil {
			log.Println("Error marshalling request payload:", err)
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}

		// Make the HTTP request to Claude API
		req, err := http.NewRequest("POST", CLAUDE_API_URL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Println("Error creating request to Claude API:", err)
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", CLAUDE_API_KEY)
		req.Header.Set("anthropic-version", "2023-06-01")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error sending request to Claude API:", err)
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		defer resp.Body.Close()

		// Read the response from Claude API
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading response from Claude API:", err)
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}

		if resp.StatusCode == http.StatusUnauthorized {
			log.Println("Invalid API Key. Please check your CLAUDE_API_KEY environment variable.")
			c.String(http.StatusUnauthorized, "Invalid API Key")
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Println("Claude API returned non-200 status:", resp.StatusCode, string(body))
			c.String(http.StatusInternalServerError, "Claude API error: "+string(body))
			return
		}

		// Parse the response to extract the content field
		var responseJson struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.Unmarshal(body, &responseJson); err != nil {
			log.Println("Error unmarshalling response body:", err)
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}

		// Extract text content
		responseText := ""
		for _, content := range responseJson.Content {
			if content.Type == "text" {
				responseText += content.Text
			}
		}

		// Return the Claude API response text
		c.String(http.StatusOK, responseText)
	})
}

// CORS middleware function
func LLMCorsMiddleware() gin.HandlerFunc {
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
