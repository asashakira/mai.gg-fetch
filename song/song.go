package song

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/asashakira/mai.gg-fetcher/utils"
)

type Song struct {
	SongID      string `json:"songID,omitempty"`
	AltKey      string `json:"altkey,omitempty"`
	Title       string `json:"title,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Bpm         string `json:"bpm,omitempty"`
	ImageUrl    string `json:"imageUrl,omitempty"`
	Version     string `json:"version,omitempty"`
	IsUtage     bool   `json:"isUtage,omitempty"`
	IsAvailable bool   `json:"isAvailable,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	DeleteDate  string `json:"deleteDate,omitempty"`
}

func (s *Song) Format() {
	s.Title = utils.RemoveNote(s.Title)
	s.Artist = utils.RemoveNote(s.Artist)
	s.Genre = utils.RemoveNote(s.Genre)
	s.Bpm = utils.RemoveNote(s.Bpm)
	s.ReleaseDate = utils.RemoveNote(s.ReleaseDate)
	s.ReleaseDate = utils.FormatDate(s.ReleaseDate)
	s.DeleteDate = utils.RemoveNote(s.DeleteDate)
	s.DeleteDate = utils.FormatDate(s.DeleteDate)

	s.AltKey = utils.CreateAltKey(s.Title, s.Artist)
}

func PrintSongs(songs []Song) {
	for _, song := range songs {
		PrintSong(song)
	}
}

func PrintSong(song Song) {
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

func DumpSongsAsJson(songs []Song) error {
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
