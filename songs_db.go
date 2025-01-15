package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// insert to db if song doesn't exist
// if not update the song with given fields
func upsertSong(song Song) (Song, error) {
	dbsong, err := getSongByAltKey(song.Title, song.Artist)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// insert if it does not exist in DB
			newSong, insertErr := insertSong(song)
			if insertErr != nil {
				return Song{}, fmt.Errorf("failed to insert song '%v': %w", song.Title, insertErr)
			}
			// return newly created song
			return newSong, nil
		}

		// other errors
		return Song{}, fmt.Errorf("failed to get song '%v': %w", song.Title, err)
	}

	// update with new fields
	song.SongID = dbsong.SongID
	song.Version = ""
	_, updateErr := updateDBSong(song)
	if updateErr != nil {
		return Song{}, fmt.Errorf("failed to update song '%v': %w", song.Title, updateErr)
	}

	return dbsong, nil
}

func getAllSongsFromDB() ([]Song, error) {
	dbURL := "http://localhost:8080/v1/songs"
	req, _ := http.NewRequest("GET", dbURL, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return []Song{}, err
	}

	if r.StatusCode != http.StatusOK {
		bodyString := string(bodyBytes)
		return []Song{}, fmt.Errorf("%v", bodyString)
	}

	songs := []Song{}
	json.Unmarshal(bodyBytes, &songs)

	return songs, nil
}

// get songs from DB using title
// may return multiple songs with same title
func getSongsByTitle(title string) ([]Song, error) {
	// Define the URL
	dbURL := fmt.Sprintf("http://localhost:8080/v1/songs/by-title/%s", url.QueryEscape(title))
	err := validateURL(dbURL)
	if err != nil {
		return []Song{}, fmt.Errorf("invalid url: %w", err)
	}

	// Create new HTTP request
	req, err := http.NewRequest("GET", dbURL, nil)
	if err != nil {
		return []Song{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	r, err := client.Do(req)
	// log.Println("GET", dbURL)
	if err != nil {
		return []Song{}, fmt.Errorf("request failed: %w", err)
	}
	defer r.Body.Close()

	// Read and process reponse body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return []Song{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// handle status code
	switch r.StatusCode {
	case http.StatusOK:
		// parse response json
		songs := []Song{}
		if err := json.Unmarshal(bodyBytes, &songs); err != nil {
			return []Song{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return songs, nil

	case http.StatusNotFound:
		// Song not found in the database
		return []Song{}, fmt.Errorf("song with title '%s' not found", title)

	default:
		return []Song{}, fmt.Errorf("error from server (status %d): %s", r.StatusCode, string(bodyBytes))
	}
}

// get song from DB using altkey
// returns one song
func getSongByAltKey(title, artist string) (Song, error) {
	// Define the URL
	altkey := createAltKey(title, artist)
	dbURL := fmt.Sprintf("http://localhost:8080/v1/songs/by-altkey/%s", url.QueryEscape(altkey))
	err := validateURL(dbURL)
	if err != nil {
		return Song{}, fmt.Errorf("invalid url: %w", err)
	}

	// Create new HTTP request
	req, err := http.NewRequest("GET", dbURL, nil)
	if err != nil {
		return Song{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	r, err := client.Do(req)
	// log.Println("GET", dbURL)
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

	case http.StatusNotFound:
		// Song not found in the database
		return Song{}, fmt.Errorf("song with title '%s' not found", title)

	default:
		return Song{}, fmt.Errorf("error from server (status %d): %s", r.StatusCode, string(bodyBytes))
	}
}

func insertSong(song Song) (Song, error) {
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
