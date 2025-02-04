package services

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

func GetUser() []User {
	resp, err := http.Get("https://my-json-server.typicode.com/uzzalcse/simple_server_with_json_placeholder/users")
	if err != nil {
		log.Printf("Failed to make HTTP request: %v", err)
		return []User{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to get users. Status code: %v", resp.StatusCode)
		return []User{}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return []User{}
	}

	var users []User
	err = json.Unmarshal(respBody, &users)
	if err != nil {
		log.Printf("Failed to unmarshal response body: %v", err)
		return []User{}
	}

	return users
}