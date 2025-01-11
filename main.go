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
			log.Println(syncErr)
		}
	}
	log.Println("fetchSongsFromAPI done")

	// scrape songs and beatmaps from gamerch
	log.Println("scraping gamerch")
	songs, beatmaps, scrapeGamerchErr := scrapeGamerch()
	check(scrapeGamerchErr)
	log.Println("done")

	// upsert songs
	log.Println("upsert songs to db...")
	dumpSongsAsJson(songs)
	for _, song := range songs {
		_, upsertErr := upsertSong(song)
		if upsertErr != nil {
			log.Println(upsertErr)
		}
	}
	log.Println("done")

	// save beatmaps to db
	log.Println("insert beatmaps to db...")
	dumpBeatmapsAsJson(beatmaps)
	for _, beatmap := range beatmaps {
		_, insertErr := upsertBeatmap(beatmap)
		if insertErr != nil {
			log.Println(insertErr)
		}
	}
	log.Println("done")
}
