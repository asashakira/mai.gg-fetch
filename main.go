package main

import (
	"log"
)

func main() {
	run()
}

func run() {
	// fetch songs from official api and save to db
	log.Println("start fetchSongsFromAPI")
	apiSongs, fetchFromAPIErr := fetchSongsFromAPI()
	check(fetchFromAPIErr)
	for _, song := range apiSongs {
		_, syncErr := upsertSong(song)
		if syncErr != nil {
			log.Fatal(syncErr)
		}
	}
	log.Println("fetchSongsFromAPI done")

	// scrape songs and beatmaps from gamerch
	log.Println("start scraping gamerch")
	songs, beatmaps, scrapeGamerchErr := scrapeGamerch()
	check(scrapeGamerchErr)
	log.Println("scraping gamerch done")

	// upsert songs
	log.Println("upsert songs to db...")
	for _, song := range songs {
		_, upsertErr := upsertSong(song)
		if upsertErr != nil {
			log.Fatal(upsertErr)
		}
	}
	log.Println("upsert songs done")

	// save beatmaps to db
	log.Println("insert beatmaps to db...")
	for _, beatmap := range beatmaps {
		_, insertErr := upsertBeatmap(beatmap)
		if insertErr != nil {
			log.Fatal(insertErr)
		}
	}
	log.Println("insert beatmaps done")
}
