package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func saveBeatmapsToDB(beatmaps []Beatmap) ([]Beatmap, error) {
	var bb []Beatmap
	for _, beatmap := range beatmaps {
		b, err := saveBeatmapToDB(beatmap)
		if err != nil {
			log.Println(beatmap.SongID, err)
			continue
		}
		bb = append(bb, b)
	}
	return bb, nil
}

func saveBeatmapToDB(beatmap Beatmap) (Beatmap, error) {
	// Defined URL
	dbURL := "http://localhost:8080/v1/beatmaps"
	err := validateURL(dbURL)
	if err != nil {
		return Beatmap{}, fmt.Errorf("invalid url: %w", err)
	}

	// Create New Request
	jsonStr, err := json.Marshal(beatmap)
	if err != nil {
		return Beatmap{}, fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	req, err := http.NewRequest("POST", dbURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return Beatmap{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return Beatmap{}, fmt.Errorf("request failed: %w", err)
	}
	defer r.Body.Close()

	// Read and process reponse body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return Beatmap{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// handle status code
	switch r.StatusCode {
	case http.StatusOK:
		// parse response json
		beatmap := Beatmap{}
		if err := json.Unmarshal(bodyBytes, &beatmap); err != nil {
			return Beatmap{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return beatmap, nil

	default:
		return Beatmap{}, fmt.Errorf("error from server (status %d): %s", r.StatusCode, string(bodyBytes))
	}
}
