package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Beatmap struct {
	SongID        string `json:"songID"`
	Difficulty    string `json:"difficulty"`
	Level         string `json:"level"`
	InternalLevel string `json:"internalLevel"`
	Type          string `json:"type"`
	TotalNotes    int32  `json:"totalNotes"`
	Tap           int32  `json:"tap"`
	Hold          int32  `json:"hold"`
	Slide         int32  `json:"slide"`
	Touch         int32  `json:"touch"`
	Break         int32  `json:"break"`
	NoteDesigner  string `json:"noteDesigner"`
	MaxDxScore    int32  `json:"maxDxScore"`
	PlayCount     int32  `json:"playCount"`
	Version       string `json:"version"`
	LastPlayedAt  string `json:"lastPlayedAt"`
}

func printBeatmaps(beatmap []Beatmap) {
	for _, beatmap := range beatmap {
		fmt.Println(beatmap)
	}
}

func getBeatmapDataFromGamerch() []Beatmap {
	beatmaps := []Beatmap{}

	songURLs := getSongURLsFromGamerch()
	for _, url := range songURLs {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
		req.Header.Set("Referer", "https://gamerch.com/maimai/545589")

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			log.Println(bodyString)
		}

		beatmap := Beatmap{}
		beatmaps = append(beatmaps, beatmap)
		time.Sleep(time.Second)
	}
	return beatmaps
}

func getBeatmapsFromDB() []Beatmap {
	dbURL := "http://localhost:8080/v1/beatmaps"
	req, _ := http.NewRequest("GET", dbURL, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Println(bodyString)

	beatmaps := []Beatmap{}
	json.Unmarshal(bodyBytes, &beatmaps)
	for _, beatmap := range beatmaps {
		fmt.Println(beatmap.Difficulty)
	}

	return beatmaps
}

func saveBeatmapsToDB(beatmaps []Beatmap) {
	// save to db
	dbURL := "http://localhost:8080/v1/beatmaps"
	for _, beatmap := range beatmaps {
		jsonStr, _ := json.Marshal(beatmap)
		req, _ := http.NewRequest("POST", dbURL, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			log.Println(beatmap.Break, bodyString)
		}
	}
}
