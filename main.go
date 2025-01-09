package main

import (
	"fmt"
	"strings"
)

func main() {
	findDupes()
	fmt.Println(".")
}

func findDupes() {
	allSongs1, err := getAllSongsFromDB()
	check(err)
	allSongs2, err := getAllSongsFromDB()
	check(err)

	for i := 0; i < len(allSongs1); i++ {
		a1 := allSongs1[i]
		for j := i + 1; j < len(allSongs2); j++ {
			a2 := allSongs2[j]

			if a1.Title == a2.Title {
				fmt.Println(a1.Title, a1.Artist)
				fmt.Println(a2.Title, a2.Artist)
				fmt.Println()
			}
		}
	}
}

func doshit2() {
	syncSongs()
	songs, _, err := scrapeGamerch()
	check(err)
	for _, song := range songs {
		s, err := getSongFromDB(song.Title, song.Artist)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// create new song if it does not exist in DB
				_, saveErr := saveSongToDB(song)
				if saveErr != nil {
					fmt.Printf("failed to save song: '%v' %v", song.Title, saveErr)
					continue
				}
				fmt.Println("new song created", song.Title)
				continue
			}
			fmt.Println("get err", song.Title, err)
			continue
		}

		s.Bpm = song.Bpm
		s.ReleaseDate = song.ReleaseDate
		_, updateErr := updateDBSong(s)
		if updateErr != nil {
			fmt.Println("update err", song.Title, updateErr)
			continue
		}
	}
}

func doshit() {
	_, beatmaps, _ := scrapeGamerch()
	dumpBeatmapsAsJson(beatmaps[0:1000])
	// saveSongsToDB(songs)
	_, err := saveBeatmapsToDB(beatmaps)
	check(err)
}
