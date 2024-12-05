package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Load environment variables
	OPEN_AI_API_KEY := os.Getenv("OPEN_AI_API_KEY")
	PASS_KEY := os.Getenv("PASS_KEY")

	if OPEN_AI_API_KEY == "" || PASS_KEY == "" {
		log.Fatal("OPEN_AI_API_KEY and PASS_KEY environment variables are required")
	}

	// Configure custom HTTP client for OpenAI with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		IdleConnTimeout:     90 * time.Second,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second, // Set a reasonable timeout
	}

	// Initialize OpenAI client with custom HTTP client
	client := openai.NewClientWithConfig(openai.ClientConfig{

		HTTPClient: httpClient,
	})

	// Set up Gin router
	router := gin.Default()

	// Enable CORS middleware
	router.Use(corsMiddleware())

	// Define the /llm endpoint
	router.POST("/llm", func(c *gin.Context) {
		// Extract prompt from JSON payload
		var jsonData struct {
			Prompt string `json:"prompt"`
		}
		if err := c.BindJSON(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}
		passhash := authHeader[7:]

		// Compute SHA256 hash of PASS_KEY
		sha256Hash := sha256.Sum256([]byte(PASS_KEY))
		HASH_KEY := hex.EncodeToString(sha256Hash[:])

		// Verify the hashed passkey
		if passhash != HASH_KEY {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		if jsonData.Prompt == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing prompt parameter"})
			return
		}

		// Log the prompt (optional)
		fmt.Println("Prompt:", jsonData.Prompt)

		// Prepare the messages for the chat completion
		messages := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: `You are an AI agent tasked with summarizing and analyzing financial information for market research.
Your response will follow the task template given to you based on the financial data given to you. Give your summarized response. You will
respond to the following prompt in a structured bullet point based format. You will also include annotations to relevant sources from the web throughout the text, attached to the important bullet points.
If the data containing the information is not relevant nor sufficient,
you may ask for more information in the response.
However, if the data containing the information is relevant to the prompt template, generate a market analysis report over the information
in accordance with the prompt template and categorize your analysis as either bullish, neutral, or bearish. Nothing more nothing less.`,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: jsonData.Prompt,
			},
		}

		// Call the OpenAI API for chat completion
		resp, err := client.CreateChatCompletion(c.Request.Context(), openai.ChatCompletionRequest{
			Model:    "gpt-4o", // Adjust model as needed
			Messages: messages,
		})

		if err != nil {
			log.Println("OpenAI error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Extract generated content from response
		generatedText := resp.Choices[0].Message.Content

		// Log the generated response (optional)
		fmt.Println("Generated Text:", generatedText)

		// Return the response as plain text
		c.String(http.StatusOK, generatedText)
	})

	// Run the server
	router.Run()
}

// CORS middleware function
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
