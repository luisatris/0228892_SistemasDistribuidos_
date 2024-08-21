package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct { // define la estructura del usuario
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Age       int    `json:"age"`
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	// revisa que el request sea valido (GET)
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	peter := User{
		Firstname: "John",
		Lastname:  "Doe",
		Age:       25,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(peter); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	// checa que el metodo sea POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "%s %s is %d years old!", user.Firstname, user.Lastname, user.Age)
}

func main() {
	http.HandleFunc("/encode", encodeHandler) // Register the encode handler
	http.HandleFunc("/decode", decodeHandler) // Register the decode handler

	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
