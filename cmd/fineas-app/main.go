package main

import (
	"fineas/api"
	"log"
	"net/http"
)

// entry point
func main() {
	certFile := "../../utils/keys/fineasapp.io.cer"
	keyFile := "../../utils/keys/fineasapp.io.key"
	go func() {
		http.Handle("/", api.CorsMiddleware(http.HandlerFunc(api.HandleQuoteRequest)))
		log.Println(http.ListenAndServeTLS(":8080", certFile, keyFile, nil))
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

	// Keep the main goroutine running to prevent the program from exiting
	select {}
}
