package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func FetchDocumentWithRetry(url string) (*goquery.Document, error) {
	for {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
		req.Header.Set("Referer", "https://gamerch.com/maimai/545589")

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer res.Body.Close()

		switch res.StatusCode {
		case http.StatusOK:
			return goquery.NewDocumentFromReader(res.Body)
		default:
			// bodyBytes, err := io.ReadAll(res.Body)
			// if err != nil {
			// 	return nil, fmt.Errorf("error reading the body: %w", err)
			// }
			retryAfter := 30 * time.Second
			fmt.Printf("error code %d: retrying after %s seconds\n", res.StatusCode, retryAfter)

			// wait to not get ip blocked
			time.Sleep(retryAfter)
		}
	}
}

func SaveHTMLToFile(html, directory, filename string) error {
	// write to file
	os.MkdirAll(directory, os.ModePerm)
	err := os.WriteFile(directory+filename, []byte(html), 0666)
	if err != nil {
		return err
	}

	return nil
}

func LoadHTMLDocument(filePath string) (*goquery.Document, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer f.Close()
	return goquery.NewDocumentFromReader(f)
}

// create alt key using song title and artist
func CreateAltKey(title, artist string) string {
	altkey := title + artist
	altkey = strings.ToLower(altkey) // all lowercase
	// altTitle = strings.Join(strings.Fields(altTitle), "") // remove all various whitespace characters
	return RemoveFromString(altkey, `[^一-龠ぁ-ゔァ-ヴーa-zA-Z0-9ａ-ｚＡ-Ｚ０-９々〆〤ヶ]+`)
}

// strips everything but numbers from a string and convert to int32
func ConvertStringToInt32(s string) int32 {
	re := regexp.MustCompile(`[0-9]+`)
	result, _ := strconv.Atoi(strings.Join(re.FindAllString(s, -1), ""))
	return int32(result)
}

// remove *n where n is an integer
// xd
func RemoveNote(s string) string {
	if strings.Contains(s, "*27") { // DECO*27
		return s
	}
	return RemoveFromString(s, `\*[0-9]+`)
}

func RemoveTM(s string) string {
	return RemoveFromString(s, `™`)
}

func RemoveFromString(input, pattern string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(input, "")
}

func FindFromString(input, pattern string) string {
	re := regexp.MustCompile(pattern)
	return re.FindString(input)
}

func FormatDate(s string) string {
	return strings.ReplaceAll(s, "/", "-")
}

// ValidateURL checks if the input string is a valid URL
func ValidateURL(input string) error {
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

func DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// FileExists checks if a file exists and is not a directory
func FileExists(filepath string) (bool, error) {
	info, err := os.Stat(filepath)
	if err == nil {
		// exists
		return !info.IsDir(), nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// parse imgSrc to determine difficulty
func GetDifficultyFromImgSrc(imgSrc string) string {
	imgName := RemoveFromString(imgSrc, `https://maimaidx.jp/maimai-mobile/img/`)
	switch imgName {
	case "diff_basic.png":
		return "basic"
	case "diff_advanced.png":
		return "advanced"
	case "diff_expert.png":
		return "expert"
	case "diff_master.png":
		return "master"
	case "diff_remaster.png":
		return "re:master"
	default:
		return ""
	}
}

// if the beatmap has Touch notes -> dx beatmap
func DetermineBeatmapType(headerText string) string {
	if strings.Contains(headerText, "Touch") {
		return "dx"
	}
	return "std"
}
