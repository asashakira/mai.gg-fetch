package main

import (
	"os"

	"github.com/asashakira/mai.gg-fetcher/maimai"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load(".env")
	segaID := os.Getenv("SEGA_ID")
	password := os.Getenv("PASSWORD")
	// Login
	m := maimai.New()
	err := m.Login(segaID, password)
	check(err)
}

func check(e error) {
	if e != nil {
		// fmt.Println(e)
		panic(e)
	}
}
