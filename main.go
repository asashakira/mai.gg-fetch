package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeSongToJson() {
	songs := getSongsFromGamerch()
	j, err := json.Marshal(songs)
	check(err)
	err = os.WriteFile("out.json", j, 0666)
	check(err)
}

func main() {
	// getBeatmapDataFromGamerch()

	fmt.Println(".")
}
