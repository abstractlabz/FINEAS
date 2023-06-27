package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/desc", descriptioService)
	fmt.Println("Server listening on port 8084...")
	log.Println(http.ListenAndServe(":8084", nil))
}

func descriptioService(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	ticker := queryParams.Get("ticker")
	fmt.Println(ticker)

}
