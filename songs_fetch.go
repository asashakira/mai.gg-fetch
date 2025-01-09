package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

func fetchSongURLsFromGamerch() ([]string, error) {
	songListURL := "https://gamerch.com/maimai/545589"
	doc, err := loadDocument(songListURL)
	if err != nil {
		return nil, err
	}

	var songURLs []string
	doc.Find(".markup.mu .mu__list--1").Each(func(i int, s *goquery.Selection) {
		url := s.Find("a").AttrOr("href", "hohoho")
		songURLs = append(songURLs, url)
	})

	return songURLs, nil
}

// fetch deleted songs title and artist
func fetchDeletedSongs() ([]Song, error) {
	deletedSongsList := "https://gamerch.com/maimai/533442"
	doc, err := loadDocument(deletedSongsList)
	if err != nil {
		return nil, err
	}

	var songs []Song
	doc.Find(".main td.mu__table--col2").Each(func(i int, s *goquery.Selection) {
		song := Song{
			Title: s.Text(),
		}
		songs = append(songs, song)
	})

	return songs, nil
}

type songsFromAPIParameters struct {
	Artist    string `json:"artist"`
	TitleKana string `json:"title_kana"`
	Genre     string `json:"catcode"`
	Comment   string `json:"comment"`
	Kanji     string `json:"kanji"`
	Title     string `json:"title"`
	ImageUrl  string `json:"image_url"`
	Release   string `json:"release"`
	Version   string `json:"version"`
}

// fetch from official api
func fetchSongsFromAPI() ([]Song, error) {
	apiURL := "https://maimai.sega.jp/data/maimai_songs.json"
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return []Song{}, fmt.Errorf("request failed: %w", err)
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	params := []songsFromAPIParameters{}
	decoder.Decode(&params)

	songs := []Song{}
	for _, p := range params {
		p = handleEdgeCases(p)

		song := Song{
			AltKey:      createAltKey(p.Title, p.Artist),
			Title:       p.Title,
			Artist:      p.Artist,
			Genre:       p.Genre,
			Bpm:         "-",
			ImageUrl:    "https://maimaidx.jp/maimai-mobile/img/Music/" + p.ImageUrl,
			Version:     versionMap[p.Version[0:3]],
			IsUtage:     p.Genre == "宴会場",
			IsAvailable: true,
			ReleaseDate: fmt.Sprintf("20%v-%v-%v", p.Release[0:2], p.Release[2:4], p.Release[4:6]),
		}
		songs = append(songs, song)
	}

	return songs, nil
}

func handleEdgeCases(p songsFromAPIParameters) songsFromAPIParameters {
	// この宴譜面だけ２つあるのでコメントで差別化
	if p.Title == "[協]青春コンプレックス" {
		suffix := "（ヒーロー級）"
		if p.Comment == "バンドメンバーを集めて楽しもう！（入門編）" {
			suffix = "（入門編）"
		}
		p.Title += suffix
	}

	// artist消されてる
	// 炎上したからっぽい
	if p.Title == "ぽっぴっぽー" {
		p.Artist = "(ラマーズP)"
	}

	// 000000 is invalid
	if p.Release == "000000" {
		p.Release = "060102"
	}

	return p
}

// fetch song data from maimaidx.jp
// pretty much only can get title
func fetchSongsFromMaimaiDxNet() []Song {
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
			Title: title,
		}
		songs = append(songs, song)
	})

	return songs
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

// get song from DB using altkey
func getSongFromDB(title, artist string) (Song, error) {
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
