package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func GetUserBySegaID(segaID, password string) (User, error) {
	// Define the URL
	dbURL := fmt.Sprintf("http://localhost:8080/v1/users/by-sega-id/%s/%s", url.QueryEscape(segaID), url.QueryEscape(password))

	// Create new HTTP request
	req, err := http.NewRequest("GET", dbURL, nil)
	if err != nil {
		return User{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("request failed: %w", err)
	}
	defer r.Body.Close()

	// Read and process reponse body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return User{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// handle status code
	switch r.StatusCode {
	case http.StatusOK:
		// parse response json
		user := User{}
		if err := json.Unmarshal(bodyBytes, &user); err != nil {
			return User{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return user, nil

	case http.StatusNotFound:
		// user not found in the database
		return User{}, fmt.Errorf("user '%s' not found", segaID)

	default:
		return User{}, fmt.Errorf("error from server (status %d): %s", r.StatusCode, string(bodyBytes))
	}
}

func GetUserScrapeMetadata(userID string) (UserScrapeMetadata, error) {
	// Define the URL
	dbURL := fmt.Sprintf("http://localhost:8080/v1/users/by-id/%s/metadata", userID)

	// Create new HTTP request
	req, err := http.NewRequest("GET", dbURL, nil)
	if err != nil {
		return UserScrapeMetadata{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return UserScrapeMetadata{}, fmt.Errorf("request failed: %w", err)
	}
	defer r.Body.Close()

	// Read and process reponse body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return UserScrapeMetadata{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// handle status code
	switch r.StatusCode {
	case http.StatusOK:
		// parse response json
		user := UserScrapeMetadata{}
		if err := json.Unmarshal(bodyBytes, &user); err != nil {
			return UserScrapeMetadata{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return user, nil

	case http.StatusNotFound:
		// user not found in the database
		return UserScrapeMetadata{}, fmt.Errorf("user scrape metadata for '%s' not found", userID)

	default:
		return UserScrapeMetadata{}, fmt.Errorf("error from server (status %d): %s", r.StatusCode, string(bodyBytes))
	}
}

func UpdateUserScrapeMetadata(u UserScrapeMetadata) (UserScrapeMetadata, error) {
	// Define the URL
	dbURL := "http://localhost:8080/v1/users/metadata"

	// Create New Request
	jsonStr, err := json.Marshal(u)
	if err != nil {
		return UserScrapeMetadata{}, fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	req, err := http.NewRequest("PATCH", dbURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return UserScrapeMetadata{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return UserScrapeMetadata{}, fmt.Errorf("request failed: %w", err)
	}
	defer r.Body.Close()

	// Read and process reponse body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return UserScrapeMetadata{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// handle status code
	switch r.StatusCode {
	case http.StatusOK:
		// parse response json
		user := UserScrapeMetadata{}
		if err := json.Unmarshal(bodyBytes, &user); err != nil {
			return UserScrapeMetadata{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return user, nil
	default:
		return UserScrapeMetadata{}, fmt.Errorf("error from server (status %d): %s", r.StatusCode, string(bodyBytes))
	}
}
