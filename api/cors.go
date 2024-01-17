package api

import "net/http"

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // This is okay for development, but use your own domain in production
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Check if it's a preflight request
		if r.Method == "OPTIONS" {
			// If so, just return a 200 status without processing further
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass the request along to the actual handler
		next.ServeHTTP(w, r)
	})
}
