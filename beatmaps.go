package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
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
	IsValid       bool   `json:"isValid"`
	LastPlayedAt  string `json:"lastPlayedAt"`
}

func printBeatmaps(beatmap []Beatmap) {
	for _, beatmap := range beatmap {
		fmt.Println(beatmap)
	}
}

func parseDifficulty(s string) (string, error) {
	colorToDifficulty := map[string]string{
		"#00ced1": "easy",
		"#98fb98": "basic",
		"#ffa500": "advanced",
		"#fa8080": "expert",
		"#ee82ee": "master",
		"#ffceff": "re:master",
		"#ff5296": "utage",
	}

	re := regexp.MustCompile(`background-color:(#[0-9a-f]+)`)
	match := re.FindStringSubmatch(s)
	if len(match) < 2 {
		return "", fmt.Errorf("failed to parse difficulty")
	}
	color := match[1]
	return colorToDifficulty[color], nil
}

func parseBeatmapRow(row *goquery.Selection, hasInternalLevel bool, beatmapType string) (Beatmap, error) {
	var beatmap Beatmap
	columns := []string{"level", "internalLevel", "totalNotes", "Tap", "Hold", "Slide", "Touch", "Break"}
	current := row.Find("th")
	for i := 0; i < len(columns); i++ {
		switch i {
		case 0: // Level
			beatmap.Level = current.Text()
		case 1: // 譜面定数
			// なければスキップ
			if !hasInternalLevel {
				continue
			}
			beatmap.InternalLevel = current.Text()
			if current.Text() == "" {
				beatmap.InternalLevel = "-"
			}
		case 2: // 総数
			beatmap.TotalNotes = convertStringToInt32(current.Text())
		case 3: // Tap
			beatmap.Tap = convertStringToInt32(current.Text())
		case 4: // Hold
			beatmap.Hold = convertStringToInt32(current.Text())
		case 5: // Slide
			beatmap.Slide = convertStringToInt32(current.Text())
		case 6: // Touch
			// standard譜面にはtouchない
			if beatmapType == "std" {
				continue
			}
			beatmap.Touch = convertStringToInt32(current.Text())
		case 7: // Break
			beatmap.Break = convertStringToInt32(current.Text())
		}
		current = current.Next()
	}

	beatmap.Difficulty, _ = parseDifficulty(row.Find("th").AttrOr("style", "yo what's up"))
	beatmap.Type = beatmapType
	beatmap.NoteDesigner = "?"
	beatmap.MaxDxScore = beatmap.TotalNotes * 3
	beatmap.PlayCount = -1
	beatmap.IsValid = true

	// beatmap validation
	isValidBeatmap := beatmap.TotalNotes > 0
	if !isValidBeatmap {
		beatmap.IsValid = false
		// return beatmap, fmt.Errorf("bad beatmap")
	}

	return beatmap, nil
}

func parseDocument(doc *goquery.Document) ([]Beatmap, error) {
	var beatmaps []Beatmap
	var songName string
	doc.Find("table").Each(func(j int, s *goquery.Selection) {
		// check top left cell to determine which table
		topLeftCell := s.Find(".mu__table--row1 .mu__table--col1").Text()

		// 基本データ
		if topLeftCell == "" && j == 0 {
			// TODO: get song data
			songName = s.Find(".mu__table--row3 .mu__table--col3").Text()
			return
		}

		// 譜面データ
		if topLeftCell == "Lv" {
			// check for missing columns
			tableHeader := s.Find("thead th").Text()
			hasInternalLevel := strings.Contains(tableHeader, "定数")
			beatmapType := "std"
			if strings.Contains(tableHeader, "Touch") {
				// has touch notes -> dx beatmap
				beatmapType = "dx"
			}

			// parse each row
			// each row represents beatmap difficulty
			s.Find("tbody tr").Each(func(j int, row *goquery.Selection) {
				beatmap, err := parseBeatmapRow(row, hasInternalLevel, beatmapType)
				if err != nil {
					// skip append on error
					log.Printf("%v %v: ", err, songName)
					return
				}
				if beatmap.Difficulty == "" || beatmap.Difficulty == "easy" || beatmap.Difficulty == "utage" {
					// skip easy and utage maps
					// TODO: support utage maps
					return
				}
				beatmap.SongID = songName
				beatmaps = append(beatmaps, beatmap)
			})
		}
	})
	return beatmaps, nil
}

func getBeatmapsFromLocal() ([]Beatmap, error) {
	beatmaps := []Beatmap{}

	for i := 0; i < 1437; i++ {
		f, err := os.Open(fmt.Sprintf("html/%v.html", i+1))
		if err != nil {
			return nil, err
		}
		defer f.Close()

		doc, err := goquery.NewDocumentFromReader(f)
		if err != nil {
			return nil, err
		}

		beatmapSet, err := parseDocument(doc)
		if err != nil {
			return nil, err
		}

		beatmaps = append(beatmaps, beatmapSet...)
	}
	return beatmaps, nil
}

func getBeatmapsFromGamerch() ([]Beatmap, error) {
	beatmaps := []Beatmap{}

	// songURLs := getSongURLsFromGamerch()
	songURLs := []string{"https://gamerch.com/maimai/533866", "https://gamerch.com/maimai/533541", "https://gamerch.com/maimai/533652", "https://gamerch.com/maimai/534105", "https://gamerch.com/maimai/533417"}
	for _, url := range songURLs {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
		req.Header.Set("Referer", "https://gamerch.com/maimai/545589")

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			// FIXME: error message
			return nil, fmt.Errorf("%v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				// FIXME: error message
				return nil, fmt.Errorf("%v", err)
			}
			bodyString := string(bodyBytes)
			log.Fatal(bodyString)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			// FIXME: error message
			return nil, fmt.Errorf("%v", err)
		}

		beatmapSet, err := parseDocument(doc)
		if err != nil {
			// FIXME: error message
			return nil, fmt.Errorf("%v", err)
		}

		beatmaps = append(beatmaps, beatmapSet...)

		time.Sleep(time.Second)
	}
	return beatmaps, nil
}

func getBeatmapsFromDB() ([]Beatmap, error) {
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

	return beatmaps, nil
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
