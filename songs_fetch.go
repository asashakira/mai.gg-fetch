package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

func fetchSongURLsFromGamerch() ([]string, error) {
	songListURL := "https://gamerch.com/maimai/545589"
	doc, err := fetchDocumentWithRetry(songListURL)
	if err != nil {
		return nil, err
	}

	var songURLs []string
	doc.Find(".markup.mu .mu__list--1").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Find("a").Attr("href")
		if !exists {
			fmt.Printf("Could not find url: %s\n", s.Text())
			return
		}
		songURLs = append(songURLs, url)
	})

	return songURLs, nil
}

// fetch deleted songs title and artist
func fetchDeletedSongs() ([]Song, error) {
	deletedSongsList := "https://gamerch.com/maimai/533442"
	doc, err := fetchDocumentWithRetry(deletedSongsList)
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
