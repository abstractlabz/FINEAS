package main

import (
	"fineas/api"
	"log"
	"net/http"
)

// entry point
func main() {
	go func() {
		http.HandleFunc("/", api.HandleQuoteRequest)
		log.Println(http.ListenAndServe(":8080", nil))
	}()

	go func() {
		http.HandleFunc("/stk", api.STKService)
		log.Println(http.ListenAndServe(":8081", nil))
	}()

	go func() {
		http.HandleFunc("/fin", api.FinService)
		log.Println(http.ListenAndServe(":8082", nil))
	}()

	go func() {
		http.HandleFunc("/news", api.NewsService)
		log.Fatal(http.ListenAndServe(":8083", nil))
	}()

	go func() {
		http.HandleFunc("/desc", api.DescriptionService)
		log.Println(http.ListenAndServe(":8084", nil))
	}()

	// Keep the main goroutine running to prevent the program from exiting
	select {}
}
