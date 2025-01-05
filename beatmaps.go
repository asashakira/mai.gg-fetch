package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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

func getBeatmapDataFromLocal() []Beatmap {
	beatmaps := []Beatmap{}

	// for i := 0; i < 1437; i++ {
	for i := 0; i < 1; i++ {
		f, err := os.Open(fmt.Sprintf("html/%v.html", i+1))
		check(err)
		defer f.Close()

		doc, err := goquery.NewDocumentFromReader(f)
		check(err)

		doc.Find("table").Each(func(j int, s *goquery.Selection) {
			// check top left cell to determine which table
			topLeftCell := s.Find(".mu__table--row1 .mu__table--col1").Text()

			// 基本データ
			if topLeftCell == "" && j == 0 {
				s.Find(".mu__table--row2 .mu__table--col2").Text()
				return
			}

			// 譜面データ
			if topLeftCell == "Lv" {
				// 定数あるか
				hasInternalLevel := strings.Contains(s.Find("thead th").Text(), "定数")

				// dx or std
				beatmapType := "std"
				if strings.Contains(s.Find("thead th").Text(), "Touch") {
					beatmapType = "dx"
				}
				s.Find("tbody tr").Each(func(j int, s *goquery.Selection) {
					beatmap := Beatmap{}
					col := []string{"level", "internalLevel", "totalNotes", "Tap", "Hold", "Slide", "Tap", "Break"}
					now := s.Find("th")
					for k := 0; k < len(col); k++ {
						switch k {
						case 0: // Level
							beatmap.Level = now.Text()
						case 1: // 譜面定数
							if !hasInternalLevel {
								continue
							}
							beatmap.InternalLevel = now.Text()
						case 2: // 総数
							tap, err := strconv.Atoi(now.Text())
							check(err)
							beatmap.Tap = int32(tap)
						case 3: // Tap
							tap, err := strconv.Atoi(now.Text())
							check(err)
							beatmap.Tap = int32(tap)
						case 4: // Hold
							hold, err := strconv.Atoi(now.Text())
							check(err)
							beatmap.Hold = int32(hold)
						case 5: // Slide
							slide, err := strconv.Atoi(now.Text())
							check(err)
							beatmap.Slide = int32(slide)
						case 6: // Touch
							if beatmapType == "std" {
								continue
							}
							touch, err := strconv.Atoi(now.Text())
							check(err)
							beatmap.Touch = int32(touch)
						case 7: // Break
							slide, err := strconv.Atoi(now.Text())
							check(err)
							beatmap.Break = int32(slide)
						}
						now = now.Next()
					}
					beatmaps = append(beatmaps, beatmap)
				})
			}
		})
	}
	return beatmaps
}

func getBeatmapDataFromGamerch() []Beatmap {
	beatmaps := []Beatmap{}

	// songURLs := getSongURLsFromGamerch()
	// シンフォ
	// oshama
	// ジャガー
	// True
	// ブリキ
	songURLs := []string{"https://gamerch.com/maimai/533866", "https://gamerch.com/maimai/533541", "https://gamerch.com/maimai/533652", "https://gamerch.com/maimai/534105", "https://gamerch.com/maimai/533417"}
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

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			panic(err)
		}
		doc.Find(".mu__table").Each(func(i int, s *goquery.Selection) {
			innerHTML, _ := s.Html()
			fmt.Println(innerHTML)
		})
		fmt.Println()

		// でらっくす譜面チェック
		// noteType := doc.Find(".mu__table--row2 .mu__table--col7").First().Text()
		// fmt.Println(noteType)

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
