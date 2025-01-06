package main

import (
	"encoding/json"
	"log"
	"net/http"
	"poc-opa-access-control-system/internal/model"
)

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Print("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if u.ID == "" {
		log.Print("Bad request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// NOTE: For simplicity, I set the user name here.
	// In a real scenario, you would fetch this information from a database or another service.
	u.Name = "John Doe"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// NOTE: The PIP has the role of returning additional information required to reference the policy, but for simplicity, it returns the received user-id as is.
	if err := json.NewEncoder(w).Encode(u); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/userinfo", handleUserInfo)
	http.ListenAndServe(":8082", nil)
}
