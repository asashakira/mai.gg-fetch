package beatmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/asashakira/mai.gg-fetcher/utils"
)

func UpsertBeatmap(beatmap Beatmap) (Beatmap, error) {
	dbBeatmap, err := GetBeatmap(beatmap.SongID, beatmap.Difficulty, beatmap.Type)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// insert if it does not exist in DB
			newBeatmap, insertErr := InsertBeatmap(beatmap)
			if insertErr != nil {
				return Beatmap{}, fmt.Errorf("failed to insert beatmap: %w", insertErr)
			}
			// return newly created song
			return newBeatmap, nil
		}

		// other errors
		return Beatmap{}, fmt.Errorf("failed to get beatmap: %w", err)
	}

	// update with new fields
	beatmap.BeatmapID = dbBeatmap.BeatmapID
	_, updateErr := UpdateDBBeatmap(beatmap)
	if updateErr != nil {
		return Beatmap{}, fmt.Errorf("failed to update beatmap: %w", updateErr)
	}

	return dbBeatmap, nil
}

// get beatmap using songID, difficulty and type
func GetBeatmap(songID, difficulty, beatmapType string) (Beatmap, error) {
	// Define the URL
	dbURL := fmt.Sprintf("http://localhost:8080/v1/beatmaps/by-song-id/%s", url.QueryEscape(songID)) // this api returns multiple beatmaps
	err := utils.ValidateURL(dbURL)
	if err != nil {
		return Beatmap{}, fmt.Errorf("invalid url: %w", err)
	}

	// Create new HTTP request
	req, err := http.NewRequest("GET", dbURL, nil)
	if err != nil {
		return Beatmap{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	r, err := client.Do(req)
	// log.Println("GET", dbURL)
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
		beatmaps := []Beatmap{}
		if err := json.Unmarshal(bodyBytes, &beatmaps); err != nil {
			return Beatmap{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		for _, beatmap := range beatmaps {
			if beatmap.Difficulty == difficulty && beatmap.Type == beatmapType {
				return beatmap, nil
			}
		}
		return Beatmap{}, fmt.Errorf("beatmap with songID '%s' difficulty '%s', type '%s' not found", songID, difficulty, beatmapType)

	case http.StatusNotFound:
		// Song not found in the database
		return Beatmap{}, fmt.Errorf("beatmap with songID '%s' not found", songID)

	default:
		return Beatmap{}, fmt.Errorf("error from server (status %d): %s", r.StatusCode, string(bodyBytes))
	}
}

// insert beatmap to db
func InsertBeatmap(beatmap Beatmap) (Beatmap, error) {
	// Define URL
	dbURL := "http://localhost:8080/v1/beatmaps"
	err := utils.ValidateURL(dbURL)
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

func UpdateDBBeatmap(beatmap Beatmap) (Beatmap, error) {
	// Define URL
	dbURL := "http://localhost:8080/v1/beatmaps"

	// Create New Request
	jsonStr, err := json.Marshal(beatmap)
	if err != nil {
		return Beatmap{}, fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	req, _ := http.NewRequest("PATCH", dbURL, bytes.NewBuffer(jsonStr))
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

func GetAllBeatmapsFromDB() ([]Beatmap, error) {
	dbURL := "http://localhost:8080/v1/beatmaps"
	req, _ := http.NewRequest("GET", dbURL, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return []Beatmap{}, err
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return []Beatmap{}, err
	}

	if r.StatusCode != http.StatusOK {
		bodyString := string(bodyBytes)
		return []Beatmap{}, fmt.Errorf("%v", bodyString)
	}

	beatmaps := []Beatmap{}
	json.Unmarshal(bodyBytes, &beatmaps)
	return beatmaps, nil
}
