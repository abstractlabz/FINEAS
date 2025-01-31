package main

import (
	"fineas/api"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// entry point
func main() {
	dataCertFile := "../../utils/keys/data/fullchain.pem"
	dataKeyFile := "../../utils/keys/data/privkey.pem"

	queryCertFile := "../../utils/keys/query/fullchain.pem"
	queryKeyFile := "../../utils/keys/query/privkey.pem"
	router := gin.Default()

	go func() {
		http.Handle("/", api.CorsMiddleware(http.HandlerFunc(api.HandleQuoteRequest)))
		log.Println(http.ListenAndServe(":8080", nil))
	}()

	go func() {
		http.Handle("/stk", api.CorsMiddleware(http.HandlerFunc(api.STKService)))
		log.Println(http.ListenAndServe(":8081", nil))
	}()

	go func() {
		http.Handle("/fin", api.CorsMiddleware(http.HandlerFunc(api.FinService)))
		log.Println(http.ListenAndServe(":8082", nil))
	}()

	go func() {
		http.Handle("/news", api.CorsMiddleware(http.HandlerFunc(api.NewsService)))
		log.Fatal(http.ListenAndServe(":8083", nil))
	}()

	go func() {
		http.Handle("/desc", api.CorsMiddleware(http.HandlerFunc(api.DescriptionService)))
		log.Println(http.ListenAndServe(":8084", nil))
	}()

	go func() {
		http.Handle("/ta", api.CorsMiddleware(http.HandlerFunc(api.TechnicalAnalysisService)))
		log.Println(http.ListenAndServe(":8089", nil))
	}()
	go func() {
		http.Handle("/ret", api.CorsMiddleware(http.HandlerFunc(api.RetrieveData)))
		log.Println(http.ListenAndServeTLS(":8035", dataCertFile, dataKeyFile, nil))
	}()
	go func() {
		http.Handle("/search", api.CorsMiddleware(http.HandlerFunc(api.SearchHandler)))
		log.Println(http.ListenAndServe(":8070", nil))
	}()
	go func() {
		api.LLMHandler(router)
		log.Fatal(router.Run(":8090"))
	}()
	go func() {
		http.Handle("/chat", api.CorsMiddleware(http.HandlerFunc(api.ChatbotQuery().ServeHTTP)))
		log.Println(http.ListenAndServeTLS(":6002", queryCertFile, queryKeyFile, nil))
	}()

	go func() {
		http.Handle("/coursechat", api.CorsMiddleware(http.HandlerFunc(api.ChatbotQuery().ServeHTTP)))
		log.Println(http.ListenAndServe(":8082", nil))
	}()

	// Keep the main goroutine running to prevent the program from exiting
	select {}
}
