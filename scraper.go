package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/asashakira/mai.gg-fetcher/beatmap"
	"github.com/asashakira/mai.gg-fetcher/song"
	"github.com/asashakira/mai.gg-fetcher/utils"
	"github.com/schollz/progressbar/v3"
)

func scrapeGamerch() ([]song.Song, []beatmap.Beatmap, error) {
	// get song urls from gamerch
	songURLs, fetchSongErr := song.FetchURLsFromGamerch()
	if fetchSongErr != nil {
		return []song.Song{}, []beatmap.Beatmap{}, fmt.Errorf("%w", fetchSongErr)
	}

	// actually scrape
	bar := progressbar.Default(int64(len(songURLs)))
	songs := []song.Song{}
	beatmaps := []beatmap.Beatmap{}
	for _, url := range songURLs {
		s, beatmapSet, err := scrapePage(url)
		if err != nil {
			return []song.Song{}, []beatmap.Beatmap{}, fmt.Errorf("%w", err)
		}
		songs = append(songs, s)
		beatmaps = append(beatmaps, beatmapSet...)

		// +1 progress
		bar.Add(1)
	}
	return songs, beatmaps, nil
}

func scrapePage(url string) (song.Song, []beatmap.Beatmap, error) {
	// check if local file exists
	filename := utils.RemoveFromString(url, "https://gamerch.com/maimai/") + ".html"
	directory := "./tmp/html/"
	filepath := directory + filename
	exists, err := utils.FileExists(filepath)
	if err != nil {
		return song.Song{}, []beatmap.Beatmap{}, fmt.Errorf("%w", err)
	}
	// if not, get it
	if !exists {
		// load gamerch song page then save to ./tmp/html/
		doc, err := utils.FetchDocumentWithRetry(url)
		if err != nil {
			return song.Song{}, []beatmap.Beatmap{}, fmt.Errorf("%w", err)
		}
		html, _ := doc.Find(".markup.mu").Html()

		err = utils.SaveHTMLToFile(html, directory, filename)
		if err != nil {
			return song.Song{}, []beatmap.Beatmap{}, fmt.Errorf("%w", err)
		}

		// wait to not get ip blocked
		time.Sleep(1 * time.Second)
	}

	// load page as goquery.Document
	doc, err := utils.LoadHTMLDocument(filepath)
	if err != nil {
		return song.Song{}, []beatmap.Beatmap{}, fmt.Errorf("%w", err)
	}

	s, beatmapSet, err := parseGamerchData(doc)
	if err != nil {
		return song.Song{}, []beatmap.Beatmap{}, fmt.Errorf("error parsing page %s: %w", url, err)
	}

	return s, beatmapSet, nil
}

// Parse the HTML Document
// Each document contains data for a single song with their beatmaps
func parseGamerchData(doc *goquery.Document) (song.Song, []beatmap.Beatmap, error) {
	var song song.Song
	var beatmaps []beatmap.Beatmap

	// parse each table
	doc.Find("table").Each(func(j int, table *goquery.Selection) {
		// check top left cell to determine which table
		topLeftCell := table.Find(".mu__table--row1 .mu__table--col1").Text()

		switch {
		// 基本データ(song data table)
		case topLeftCell == "" && j == 0:
			s, err := handleSongTable(table)
			if err != nil {
				log.Printf("handle song table error: %v", err)
				return
			}
			song = s

		// 譜面データ(beatmap data table)
		case topLeftCell == "Lv":
			b, err := handleBeatmapTable(table, song)
			if err != nil {
				log.Printf("handle beatmap table error: %v", err)
				return
			}
			beatmaps = append(beatmaps, b...)
		}
	})

	return song, beatmaps, nil
}

func handleSongTable(table *goquery.Selection) (song.Song, error) {
	gamerchSong, parseSongTableErr := parseSongTable(table)
	if parseSongTableErr != nil {
		return song.Song{}, fmt.Errorf("parse song table error: %v", parseSongTableErr)
	}

	// get song from DB for SongID
	var s song.Song
	var getSongErr error
	s, getSongErr = song.GetSongByAltKey(gamerchSong.Title, gamerchSong.Artist)
	if getSongErr != nil {
		if strings.Contains(getSongErr.Error(), "not found") {
			// create new song if it does not exist in DB
			newSong, insertErr := song.InsertSong(gamerchSong)
			if insertErr != nil {
				return song.Song{}, fmt.Errorf("failed to insert song '%v': %w", gamerchSong.Title, insertErr)
			}
			return newSong, nil
		}
		// other errors
		return song.Song{}, fmt.Errorf("failed to get song '%v': %w", gamerchSong.Title, getSongErr)
	}

	return s, nil
}

// parse song table
func parseSongTable(table *goquery.Selection) (song.Song, error) {
	var genre, title, artist, bpm, releaseDate, deleteDate, version string

	table.Find(`tr`).Each(func(i int, row *goquery.Selection) {
		header := row.Find(`.mu__table--col2`).Text()
		switch header {
		case "ジャンル":
			genre = row.Find(".mu__table--col3").Text()
		case "タイトル":
			title = row.Find(".mu__table--col3").Text()
		case "アーティスト":
			artist = row.Find(".mu__table--col3").Text()
		case "BPM":
			bpm = row.Find(".mu__table--col3").Text()
		case "配信日":
			releaseDate = row.Find(".mu__table--col3").Text()
		case "削除日":
			deleteDate = row.Find(".mu__table--col3").Text()
		case "バージョン":
			version = row.Find(".mu__table--col3").Text()
		}
	})

	// ignore everything after release date
	releaseDate = utils.FindFromString(releaseDate, `^[/0-9]+`)

	s := song.Song{
		AltKey:      utils.CreateAltKey(title, artist),
		Title:       title,
		Artist:      artist,
		Genre:       genre,
		Bpm:         bpm,
		Version:     version,
		ReleaseDate: releaseDate,
		DeleteDate:  deleteDate,
	}
	s.Format()
	return s, nil
}

func handleBeatmapTable(table *goquery.Selection, s song.Song) ([]beatmap.Beatmap, error) {
	var beatmaps []beatmap.Beatmap

	// check header for missing data
	headerText := table.Find("thead th").Text()
	hasInternalLevel := strings.Contains(headerText, "定数")
	beatmapType := utils.DetermineBeatmapType(headerText)

	// parse each row
	// each table row represents beatmap difficulty
	table.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
		b, err := parseBeatmapRow(row, hasInternalLevel, beatmapType)
		if err != nil {
			// log.Printf("parse beatmap row error for song '%s': %v", s.Title, err)
		}

		if b.Difficulty == "" || b.Difficulty == "easy" || b.Difficulty == "utage" {
			// skip easy and utage maps
			// TODO: support utage maps
			return
		}

		// set SongID
		b.SongID = s.SongID
		beatmaps = append(beatmaps, b)
	})

	return beatmaps, nil
}

// parse each row of beatmaps table
func parseBeatmapRow(row *goquery.Selection, hasInternalLevel bool, beatmapType string) (beatmap.Beatmap, error) {
	var b beatmap.Beatmap
	columns := []string{"level", "internalLevel", "totalNotes", "Tap", "Hold", "Slide", "Touch", "Break"}
	current := row.Find("th")
	for i := 0; i < len(columns); i++ {
		switch i {
		case 0: // Level
			b.Level = current.Text()
		case 1: // 譜面定数
			if !hasInternalLevel { // 定数列がなければskip
				continue
			}
			// 定数列があっても空だったら無視
			internalLevelString := current.Text()
			if internalLevelString != "" {
				internalLevel, parseErr := strconv.ParseFloat(internalLevelString, 64)
				if parseErr != nil {
					return beatmap.Beatmap{}, fmt.Errorf("failed to parse internal level: %w", parseErr)
				}
				b.InternalLevel = internalLevel
			}
		case 2: // 総数
			b.TotalNotes = utils.ConvertStringToInt32(current.Text())
		case 3: // Tap
			b.Tap = utils.ConvertStringToInt32(current.Text())
		case 4: // Hold
			b.Hold = utils.ConvertStringToInt32(current.Text())
		case 5: // Slide
			b.Slide = utils.ConvertStringToInt32(current.Text())
		case 6: // Touch
			// standard譜面にはtouchない
			if beatmapType == "std" {
				continue
			}
			b.Touch = utils.ConvertStringToInt32(current.Text())
		case 7: // Break
			b.Break = utils.ConvertStringToInt32(current.Text())
		}
		current = current.Next()
	}

	b.Difficulty, _ = beatmap.ParseDifficulty(row.Find("th").AttrOr("style", "yo what's up"))
	b.Type = beatmapType
	// TODO: get NoteDesigner
	b.NoteDesigner = "?"
	b.MaxDxScore = b.TotalNotes * 3

	// beatmap validation
	b.IsValid = true
	isValidBeatmap := b.TotalNotes > 0
	if !isValidBeatmap {
		b.IsValid = false
	}
	return b, nil
}
