package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Song struct {
	SongID      string `json:"songID"`
	AltKey      string `json:"altkey"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Genre       string `json:"genre"`
	Bpm         string `json:"bpm"`
	ImageUrl    string `json:"imageUrl"`
	Version     string `json:"version"`
	IsUtage     bool   `json:"isUtage"`
	IsAvailable bool   `json:"isAvailable"`
	ReleaseDate string `json:"releaseDate"`
	DeleteDate  string `json:"deleteDate"`
}

func (s *Song) Format() {
	s.Title = removeNote(s.Title)
	s.Artist = removeNote(s.Artist)
	s.Genre = removeNote(s.Genre)
	s.ReleaseDate = removeNote(s.ReleaseDate)
	s.ReleaseDate = formatDate(s.ReleaseDate)
	s.DeleteDate = removeNote(s.DeleteDate)
	s.DeleteDate = formatDate(s.DeleteDate)

	s.AltKey = createAltKey(s.Title, s.Artist)
}

func syncSongs() {
	songs, err := fetchSongsFromAPI()
	check(err)
	for _, song := range songs {
		_, err := syncSong(song)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func syncSong(song Song) (Song, error) {
	_, err := getSongFromDB(song.Title, song.Artist)
	if err == nil {
		// song already exists
		return Song{}, nil
	}

	if strings.Contains(err.Error(), "not found") {
		// create new song if it does not exist in DB
		_, saveErr := saveSongToDB(song)
		if saveErr != nil {
			return Song{}, fmt.Errorf("failed to save song: '%v' %w", song.Title, saveErr)
		}
		return Song{}, nil
	}

	return Song{}, fmt.Errorf("failed to get song: %w", err)
}

func printSongs(songs []Song) {
	for _, song := range songs {
		printSong(song)
	}
}

func printSong(song Song) {
	fmt.Println("SongID:  ", song.SongID)
	fmt.Println("Title:   ", song.Title)
	fmt.Println("Artist:  ", song.Artist)
	fmt.Println("Genre:   ", song.Genre)
	fmt.Println("Bpm:     ", song.Bpm)
	fmt.Println("ImageURL:", song.ImageUrl)
	fmt.Println("Version: ", song.Version)
	fmt.Println("RelDate: ", song.ReleaseDate)
	fmt.Println()
}

func dumpSongsAsJson(songs []Song) error {
	jsonByte, err := json.Marshal(songs)
	if err != nil {
		return err
	}

	// write to file
	os.MkdirAll("./tmp/json/", os.ModePerm)
	err = os.WriteFile("./tmp/json/songs.json", jsonByte, 0666)
	if err != nil {
		return err
	}
	return nil
}

var versionMap = map[string]string{
	"000": "",
	"100": "maimai",
	"110": "maimai PLUS",
	"120": "GreeN",
	"130": "GreeN PLUS",
	"140": "ORANGE",
	"150": "ORANGE PLUS",
	"160": "PiNK",
	"170": "PiNK PLUS",
	"180": "MURASAKi",
	"185": "MURASAKi PLUS",
	"190": "MiLK",
	"195": "MiLK PLUS",
	"199": "FiNALE",
	"200": "maimaiでらっくす",
	"205": "maimaiでらっくす PLUS",
	"210": "Splash",
	"215": "Splash PLUS",
	"220": "UNiVERSE",
	"225": "UNiVERSE PLUS",
	"230": "FESTiVAL",
	"235": "FESTiVAL PLUS",
	"240": "BUDDiES",
	"245": "BUDDiES PLUS",
	"250": "PRiSM",
}
