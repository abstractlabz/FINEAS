package serviceauth

import (
	"net/http"
	"strings"
)

func ServiceAuthMiddleware(w http.ResponseWriter, r *http.Request, eventSequenceArray []string, passHash string) bool {

	// Get the Authorization header value
	authHeader := r.Header.Get("Authorization")

	// Check if the Authorization header is present and starts with "Bearer "
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		eventSequenceArray = append(eventSequenceArray, "passhash unauthorized \n")
		return false
	}

	// Extract the token from the Authorization header
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Perform token validation (e.g., check if it's a valid token)
	if token != passHash {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		eventSequenceArray = append(eventSequenceArray, "passhash unauthorized \n")
		return false
	}

	eventSequenceArray = append(eventSequenceArray, "passhash passed \n")
	return true

}
