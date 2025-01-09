package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func scrapeGamerch() ([]Song, []Beatmap, error) {
	dir := "./tmp/html/"
	songURLs, fetchSongErr := fetchSongURLsFromGamerch()
	if fetchSongErr != nil {
		return []Song{}, []Beatmap{}, fmt.Errorf("%w", fetchSongErr)
	}
	exists, err := dirExists(dir)
	if err != nil {
		return []Song{}, []Beatmap{}, fmt.Errorf("%w", err)
	}

	if !exists {
		err = saveGamerchHTML(songURLs)
		if err != nil {
			return []Song{}, []Beatmap{}, fmt.Errorf("%w", err)
		}
	}

	songs := []Song{}
	beatmaps := []Beatmap{}
	for i := range songURLs {
		doc, err := loadHTML(fmt.Sprintf("%s%04d.html", dir, i+1))
		if err != nil {
			return []Song{}, []Beatmap{}, fmt.Errorf("%w", err)
		}

		song, beatmapSet, err := parseDocument(doc)
		if err != nil {
			return []Song{}, []Beatmap{}, fmt.Errorf("%w", err)
		}

		songs = append(songs, song)
		beatmaps = append(beatmaps, beatmapSet...)
	}
	return songs, beatmaps, nil
}

// load gamerch song document then save to ./tmp/html/
func saveGamerchHTML(songURLs []string) error {
	for i, url := range songURLs {
		doc, err := loadDocument(url)
		if err != nil {
			return err
		}
		html, _ := doc.Find(".markup.mu").Html()

		saveHTML(html, "./tmp/html", fmt.Sprintf("%04d.html", i+1))

		// progress
		fmt.Println(i+1, "/", len(songURLs))

		// wait to not get ip blocked
		time.Sleep(time.Second)
	}
	return nil
}

// Parse the HTML Document
// Each document contains data for a single song with their beatmaps
func parseDocument(doc *goquery.Document) (Song, []Beatmap, error) {
	var song Song
	var beatmaps []Beatmap

	// parse each table
	doc.Find("table").Each(func(j int, table *goquery.Selection) {
		// check top left cell to determine which table
		topLeftCell := table.Find(".mu__table--row1 .mu__table--col1").Text()

		switch {
		case topLeftCell == "" && j == 0: // 基本データ(song data table)
			gamerchSong, err := parseSongTable(table)
			if err != nil {
				log.Printf("parse song table error: %v", err)
			}
			// get song from DB to get SongID for beatmap creation
			s, err := getSongFromDB(gamerchSong.Title, gamerchSong.Artist)
			if err != nil {
				log.Println(gamerchSong.Title, err)
			}
			song = s

		case topLeftCell == "Lv": // 譜面データ(beatmap data table)
			headerText := table.Find("thead th").Text()
			hasInternalLevel := strings.Contains(headerText, "定数")
			beatmapType := determineBeatmapType(headerText)

			// parse each row
			// each table row represents beatmap difficulty
			table.Find("tbody tr").Each(func(j int, row *goquery.Selection) {
				beatmap, err := parseBeatmapRow(row, hasInternalLevel, beatmapType)
				// skip append on error
				if err != nil {
					log.Printf("%v %v: ", err, song.Title)
					return
				}

				// skip easy and utage maps
				if beatmap.Difficulty == "" || beatmap.Difficulty == "easy" || beatmap.Difficulty == "utage" {
					// TODO: support utage maps
					return
				}

				// set SongID
				beatmap.SongID = song.SongID
				beatmaps = append(beatmaps, beatmap)
			})

		}
	})
	return song, beatmaps, nil
}

// parse song table
func parseSongTable(table *goquery.Selection) (Song, error) {
	genre := table.Find(".mu__table--row2 .mu__table--col3").Text()
	title := table.Find(".mu__table--row3 .mu__table--col3").Text()
	artist := table.Find(".mu__table--row4 .mu__table--col3").Text()
	bpm := table.Find(".mu__table--row5 .mu__table--col3").Text()
	releaseDate := table.Find(".mu__table--row6 .mu__table--col3").Text()
	version := table.Find(".mu__table--row7 .mu__table--col3").Text()
	image_url := ""

	releaseDate = getFromString(releaseDate, `^[/0-9]+`)

	song := Song{
		AltKey:      createAltKey(title, artist),
		Title:       title,
		Artist:      artist,
		Genre:       genre,
		Bpm:         bpm,
		ImageUrl:    image_url,
		Version:     version,
		ReleaseDate: releaseDate,
	}
	song.Format()
	return song, nil
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
			internalLevelString := current.Text()
			beatmap.InternalLevel, _ = strconv.ParseFloat(internalLevelString, 64)
			if current.Text() == "" {
				beatmap.InternalLevel = -1
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
	// TODO: get NoteDesigner
	beatmap.NoteDesigner = "?"
	beatmap.MaxDxScore = beatmap.TotalNotes * 3

	// beatmap validation
	beatmap.IsValid = true
	isValidBeatmap := beatmap.TotalNotes > 0
	if !isValidBeatmap {
		beatmap.IsValid = false
	}
	return beatmap, nil
}
