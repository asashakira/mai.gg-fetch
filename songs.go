package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

type Song struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Genre    string `json:"genre"`
	Bpm      string `json:"bpm"`
	ImageURL string `json:"imageURL"`
}

func printSongs(songs []Song) {
	for _, song := range songs {
		fmt.Println("Title:", song.Title)
		fmt.Println("Artist:", song.Artist)
		fmt.Println("Genre:", song.Genre)
		fmt.Println("Bpm:", song.Bpm)
		fmt.Println("ImageURL:", song.ImageURL)
		fmt.Println()
	}
}

func getSongsFromAPI() []Song {
	apiURL := "https://maimai.sega.jp/data/maimai_songs.json"
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	type parameters struct {
		Artist   string `json:"artist"`
		CatCode  string `json:"catcode"`
		Title    string `json:"title"`
		ImageURL string `json:"image_url"`
	}
	decoder := json.NewDecoder(r.Body)
	params := []parameters{}
	decoder.Decode(&params)

	songs := []Song{}
	for _, p := range params {
		song := Song{
			Title:    p.Title,
			Artist:   p.Artist,
			Genre:    p.CatCode,
			ImageURL: "https://maimaidx.jp/maimai-mobile/img/Music/" + p.ImageURL,
		}
		songs = append(songs, song)
	}

	return songs
}

func scrapeSongsFromMaimaiDxNet() []Song {
	godotenv.Load(".env")

	segaID := os.Getenv("SEGA_ID")
	password := os.Getenv("PASSWORD")
	m := New()
	err := m.Login(segaID, password)
	if err != nil {
		log.Println("maimai Login error: ", err)
		panic(err)
	}

	res, err := m.HTTPClient.Get(maimaiURL + "/record/musicGenre/search/?genre=99&diff=3")
	if err != nil {
		log.Println("GET error: ", err)
		panic(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		panic(err)
	}

	songs := []Song{}
	doc.Find(".music_master_score_back").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".music_name_block").Text()
		song := Song{
			Title:    title,
			Artist:   "",
			Genre:    "",
			Bpm:      "?",
			ImageURL: "",
		}
		songs = append(songs, song)
	})

	return songs
}

func getSongsFromDB() []Song {
	dbURL := "http://localhost:8080/v1/songs"
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

	songs := []Song{}
	json.Unmarshal(bodyBytes, &songs)
	for _, song := range songs {
		fmt.Println(song.Title)
	}

	return songs
}

func saveSongsToDB(songs []Song) {
	// save to db
	dbURL := "http://localhost:8080/v1/songs"
	for _, song := range songs {
		jsonStr, _ := json.Marshal(song)
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
			log.Println(song.Title, bodyString)
		}
	}
}
