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
	// get song urls from gamerch
	songURLs, fetchSongErr := fetchSongURLsFromGamerch()
	if fetchSongErr != nil {
		return []Song{}, []Beatmap{}, fmt.Errorf("%w", fetchSongErr)
	}

	// actually scrape
	songs := []Song{}
	beatmaps := []Beatmap{}
	for i, url := range songURLs {
		song, beatmapSet, err := scrapePage(url)
		if err != nil {
			return []Song{}, []Beatmap{}, fmt.Errorf("%w", err)
		}
		songs = append(songs, song)
		beatmaps = append(beatmaps, beatmapSet...)

		// progress
		fmt.Println(i+1, "/", len(songURLs))
	}
	return songs, beatmaps, nil
}

func scrapePage(url string) (Song, []Beatmap, error) {
	// check if local file exists
	filename := makeFilenameFromURL(url)
	directory := "./tmp/html/"
	filepath := directory + filename
	exists, err := fileExists(filepath)
	if err != nil {
		return Song{}, []Beatmap{}, fmt.Errorf("%w", err)
	}
	// if not, get it
	if !exists {
		// load gamerch song page then save to ./tmp/html/
		doc, err := fetchDocumentWithRetry(url)
		if err != nil {
			return Song{}, []Beatmap{}, fmt.Errorf("%w", err)
		}
		html, _ := doc.Find(".markup.mu").Html()

		err = saveHTMLToFile(html, directory, filename)
		if err != nil {
			return Song{}, []Beatmap{}, fmt.Errorf("%w", err)
		}

		// wait to not get ip blocked
		time.Sleep(1 * time.Second)
	}

	// load page as goquery.Document
	doc, err := loadHTMLDocument(filepath)
	if err != nil {
		fmt.Println(err)
		// return []Song{}, []Beatmap{}, fmt.Errorf("%w", err)
	}

	song, beatmapSet, err := parseGamerchData(doc)
	if err != nil {
		return Song{}, []Beatmap{}, fmt.Errorf("error parsing page %s: %w", url, err)
	}

	return song, beatmapSet, nil
}

// Parse the HTML Document
// Each document contains data for a single song with their beatmaps
func parseGamerchData(doc *goquery.Document) (Song, []Beatmap, error) {
	var song Song
	var beatmaps []Beatmap

	// parse each table
	doc.Find("table").Each(func(j int, table *goquery.Selection) {
		// check top left cell to determine which table
		topLeftCell := table.Find(".mu__table--row1 .mu__table--col1").Text()

		switch {
		// 基本データ(song data table)
		case topLeftCell == "" && j == 0:
			parsedSong, err := handleSongTable(table)
			if err != nil {
				log.Printf("handle song table error: %v", err)
				return
			}
			song = parsedSong

		// 譜面データ(beatmap data table)
		case topLeftCell == "Lv":
			parsedBeatmaps, err := handleBeatmapTable(table, song)
			if err != nil {
				log.Printf("handle beatmap table error: %v", err)
				return
			}
			beatmaps = append(beatmaps, parsedBeatmaps...)
		}
	})

	return song, beatmaps, nil
}

func handleSongTable(table *goquery.Selection) (Song, error) {
	gamerchSong, err := parseSongTable(table)
	if err != nil {
		return Song{}, fmt.Errorf("parse song table error: %v", err)
	}

	// get song from DB for SongID
	// add if doesn't exist
	song, syncErr := upsertSong(gamerchSong)
	if syncErr != nil {
		return Song{}, fmt.Errorf("failed to upsert song: %w", err)
	}

	return song, nil
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

	// ignore everything after release date
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

func handleBeatmapTable(table *goquery.Selection, song Song) ([]Beatmap, error) {
	var beatmaps []Beatmap

	// check header for missing data
	headerText := table.Find("thead th").Text()
	hasInternalLevel := strings.Contains(headerText, "定数")
	beatmapType := determineBeatmapType(headerText)

	// parse each row
	// each table row represents beatmap difficulty
	table.Find("tbody tr").Each(func(j int, row *goquery.Selection) {
		beatmap, err := parseBeatmapRow(row, hasInternalLevel, beatmapType)
		if err != nil {
			log.Printf("parse beatmap row error for song '%s': %v", song.Title, err)
			return
		}

		if beatmap.Difficulty == "" || beatmap.Difficulty == "easy" || beatmap.Difficulty == "utage" {
			// skip easy and utage maps
			// TODO: support utage maps
			return
		}

		// set SongID
		beatmap.SongID = song.SongID
		beatmaps = append(beatmaps, beatmap)
	})

	return beatmaps, nil
}

// parse each row of beatmaps table
func parseBeatmapRow(row *goquery.Selection, hasInternalLevel bool, beatmapType string) (Beatmap, error) {
	var beatmap Beatmap
	columns := []string{"level", "internalLevel", "totalNotes", "Tap", "Hold", "Slide", "Touch", "Break"}
	current := row.Find("th")
	for i := 0; i < len(columns); i++ {
		switch i {
		case 0: // Level
			beatmap.Level = current.Text()
		case 1: // 譜面定数
			if !hasInternalLevel {     // 定数列がなければskip
				continue
			}
			// 定数列があっても空だったら無視
			internalLevelString := current.Text()
			if internalLevelString != "" {
				internalLevel, parseErr := strconv.ParseFloat(internalLevelString, 64)
				if parseErr != nil {
					return Beatmap{}, fmt.Errorf("failed to parse internal level: %w", parseErr)
				}
				beatmap.InternalLevel = internalLevel
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
