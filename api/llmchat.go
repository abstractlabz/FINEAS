package api

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

func LLMHandler(router *gin.Engine) {
	// Load environment variables
	OPEN_AI_API_KEY := os.Getenv("OPEN_AI_API_KEY")
	PASS_KEY := os.Getenv("PASS_KEY")

	if OPEN_AI_API_KEY == "" || PASS_KEY == "" {
		log.Fatal("OPEN_AI_API_KEY and PASS_KEY environment variables are required")
	}

	// Initialize OpenAI client with custom HTTP client
	client := openai.NewClient(OPEN_AI_API_KEY)

	// Enable CORS middleware
	router.Use(corsMiddleware())

	// Define the /llm endpoint
	router.POST("/llm", func(c *gin.Context) {
		// Extract prompt from JSON payload
		var jsonData struct {
			Prompt string `json:"prompt"`
		}
		if err := c.BindJSON(&jsonData); err != nil {
			log.Println("Error binding JSON:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		log.Println("Received prompt:", jsonData.Prompt)

		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Println("Unauthorized access attempt with authHeader:", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}
		passhash := authHeader[7:]

		// Compute SHA256 hash of PASS_KEY
		sha256Hash := sha256.Sum256([]byte(PASS_KEY))
		HASH_KEY := hex.EncodeToString(sha256Hash[:])

		// Verify the hashed passkey
		if passhash != HASH_KEY {
			log.Println("Unauthorized access attempt with passhash:", passhash)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		if jsonData.Prompt == "" {
			log.Println("Missing prompt parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing prompt parameter"})
			return
		}

		// Log the prompt (optional)
		log.Println("Prompt:", jsonData.Prompt)

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

		log.Println("Prepared messages for OpenAI API:", messages)

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
		log.Println("Generated Text:", generatedText)

		// Return the response as plain text
		c.String(http.StatusOK, generatedText)
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
