package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func saveSongsToDB(songs []Song) ([]Song, error) {
	var ss []Song
	for _, song := range songs {
		s, err := syncSong(song)
		if err != nil {
			log.Println(err)
			continue
		}
		ss = append(ss, s)
	}
	return ss, nil
}

func saveSongToDB(song Song) (Song, error) {
	// Define URL
	dbURL := "http://localhost:8080/v1/songs"
	err := validateURL(dbURL)
	if err != nil {
		return Song{}, fmt.Errorf("invalid url: %w", err)
	}

	// Create New Request
	jsonStr, err := json.Marshal(song)
	if err != nil {
		return Song{}, fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	req, err := http.NewRequest("POST", dbURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return Song{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return Song{}, fmt.Errorf("request failed: %w", err)
	}
	defer r.Body.Close()

	// Read and process reponse body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return Song{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// handle status code
	switch r.StatusCode {
	case http.StatusOK:
		// parse response json
		song := Song{}
		if err := json.Unmarshal(bodyBytes, &song); err != nil {
			return Song{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return song, nil

	default:
		return Song{}, fmt.Errorf("error from server (status %d): %s", r.StatusCode, string(bodyBytes))
	}
}

func updateDBSong(song Song) (Song, error) {
	// Define URL
	dbURL := "http://localhost:8080/v1/songs"

	// Create New Request
	jsonStr, err := json.Marshal(song)
	if err != nil {
		return Song{}, fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	req, _ := http.NewRequest("PATCH", dbURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return Song{}, fmt.Errorf("request failed: %w", err)
	}
	defer r.Body.Close()

	// Read and process reponse body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return Song{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// handle status code
	switch r.StatusCode {
	case http.StatusOK:
		// parse response json
		song := Song{}
		if err := json.Unmarshal(bodyBytes, &song); err != nil {
			return Song{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return song, nil

	default:
		return Song{}, fmt.Errorf("error from server (status %d): %s", r.StatusCode, string(bodyBytes))
	}
}
