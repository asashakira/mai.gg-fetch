package song

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/asashakira/mai.gg-fetcher/utils"
)

func FetchURLsFromGamerch() ([]string, error) {
	songListURL := "https://gamerch.com/maimai/545589"
	doc, err := utils.FetchDocumentWithRetry(songListURL)
	if err != nil {
		return nil, err
	}

	var songURLs []string
	doc.Find(".markup.mu .mu__list--1").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Find("a").Attr("href")
		if !exists {
			fmt.Printf("Could not find url: %s\n", s.Text())
			return
		}
		songURLs = append(songURLs, url)
	})

	return songURLs, nil
}

// fetch deleted songs title and artist
func FetchDeletedSongs() ([]Song, error) {
	deletedSongsList := "https://gamerch.com/maimai/533442"
	doc, err := utils.FetchDocumentWithRetry(deletedSongsList)
	if err != nil {
		return nil, err
	}

	var songs []Song
	doc.Find(".main td.mu__table--col2").Each(func(i int, s *goquery.Selection) {
		song := Song{
			Title: s.Text(),
		}
		songs = append(songs, song)
	})

	return songs, nil
}
