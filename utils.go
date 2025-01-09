package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func check(e error) {
	if e != nil {
		// fmt.Println(e)
		panic(e)
	}
}

func loadDocument(url string) (*goquery.Document, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
	req.Header.Set("Referer", "https://gamerch.com/maimai/545589")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		// FIXME: error message
		return nil, fmt.Errorf("%v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			// FIXME: error message
			return nil, fmt.Errorf("%v", err)
		}
		bodyString := string(bodyBytes)
		log.Fatal(bodyString)
	}
	return goquery.NewDocumentFromReader(res.Body)
}

func loadHTML(filePath string) (*goquery.Document, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return goquery.NewDocumentFromReader(f)
}

// create alt key using song title and artist
func createAltKey(title, artist string) string {
	altkey := title + artist
	altkey = strings.ToLower(altkey) // all lowercase
	// altTitle = strings.Join(strings.Fields(altTitle), "") // remove all various whitespace characters
	return removeFromString(altkey, `[^一-龠ぁ-ゔァ-ヴーa-zA-Z0-9ａ-ｚＡ-Ｚ０-９々〆〤ヶ]+`)
}

func convertStringToInt32(s string) int32 {
	re := regexp.MustCompile(`[0-9]+`)
	result, _ := strconv.Atoi(strings.Join(re.FindAllString(s, -1), ""))
	return int32(result)
}

func removeNote(s string) string {
	if strings.Contains(s, "*27") { // DECO*27
		return s
	}
	return removeFromString(s, `\*[0-9]+`)
}

func removeTM(s string) string {
	return removeFromString(s, `™`)
}

func removeFromString(input, pattern string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(input, "")
}

func getFromString(input, pattern string) string {
	re := regexp.MustCompile(pattern)
	return re.FindString(input)
}

func formatDate(s string) string {
	return strings.ReplaceAll(s, "/", "-")
}

// validateURL checks if the input string is a valid URL
func validateURL(input string) error {
	parsedURL, err := url.ParseRequestURI(input)
	if err != nil {
		return errors.New("invalid URL format")
	}

	// // Ensure the URL has a valid scheme (e.g., http or https)
	// if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
	// 	return errors.New("URL must have http or https scheme")
	// }

	// Ensure the URL has a host
	if parsedURL.Host == "" {
		return errors.New("URL must have a host")
	}

	return nil
}

func saveHTML(html, directory, filename string) error {
	// write to file
	os.MkdirAll(directory, os.ModePerm)
	err := os.WriteFile(fmt.Sprintf("%s%s", directory, filename), []byte(html), 0666)
	if err != nil {
		return err
	}

	return nil
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}
