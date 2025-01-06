package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func convertStringToInt32(s string) int32 {
	re := regexp.MustCompile(`[0-9]+`)
	result, _ := strconv.Atoi(strings.Join(re.FindAllString(s, -1), ""))
	return int32(result)
}

func doshit() {
	songURLs := getSongURLsFromGamerch()
	for i, url := range songURLs {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
		req.Header.Set("Referer", "https://gamerch.com/maimai/545589")

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			log.Println(bodyString)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			panic(err)
		}
		var html string
		doc.Find(".mu__table").Each(func(_ int, s *goquery.Selection) {
			innerHTML, _ := s.Html()
			html += innerHTML
		})
		err = os.WriteFile(fmt.Sprintf("%v.html", i+1), []byte(html), 0666)
		check(err)

		fmt.Println(i+1, "/", len(songURLs))
		time.Sleep(time.Second)
	}
}

func writeSongToJson() {
	songs := getSongsFromGamerch()
	j, err := json.Marshal(songs)
	check(err)
	err = os.WriteFile("out.json", j, 0666)
	check(err)
}
