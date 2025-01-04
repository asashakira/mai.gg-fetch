package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	m := New()
	segaID := os.Getenv("SEGAID")
	password := os.Getenv("PASSWORD")
	err := m.Login(segaID, password)
	if err != nil {
		log.Println("maimai Login error: ", err)
		return
	}

	res, err := m.HttpClient.Get(maimaiUrl + "/playerData")
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	rating, _ := strconv.Atoi(doc.Find(".rating_block").Text())
	fmt.Println(rating)
}
