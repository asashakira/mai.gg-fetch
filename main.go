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
	Creator  string `json:"creator"`
	Genre    string `json:"genre"`
	Bpm      int32  `json:"bpm"`
	ImageUrl string `json:"imageUrl"`
}

func main() {
	godotenv.Load(".env")

	segaID := os.Getenv("SEGA_ID")
	password := os.Getenv("PASSWORD")
	m := New()
	err := m.Login(segaID, password)
	if err != nil {
		log.Println("maimai Login error: ", err)
		panic(err)
	}

	res, err := m.HttpClient.Get(maimaiUrl + "/record/musicGenre/search/?genre=99&diff=3")
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
			Creator:  "",
			Genre:    "",
			Bpm:      0,
			ImageUrl: "",
		}
		songs = append(songs, song)
	})
	for _, song := range songs {
		fmt.Println(song.Title)
	}

	// save to db
	url := "http://localhost:8080/v1/songs"
	for _, song := range songs {
		jsonStr, _ := json.Marshal(song)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
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
			log.Println(bodyString)
		}
	}
}
