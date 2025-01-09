package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
)

type Beatmap struct {
	SongID        string  `json:"songID"`
	Difficulty    string  `json:"difficulty"`
	Level         string  `json:"level"`
	InternalLevel float64 `json:"internalLevel"`
	Type          string  `json:"type"`
	TotalNotes    int32   `json:"totalNotes"`
	Tap           int32   `json:"tap"`
	Hold          int32   `json:"hold"`
	Slide         int32   `json:"slide"`
	Touch         int32   `json:"touch"`
	Break         int32   `json:"break"`
	NoteDesigner  string  `json:"noteDesigner"`
	MaxDxScore    int32   `json:"maxDxScore"`
	IsValid       bool    `json:"isValid"`
	LastPlayedAt  string  `json:"lastPlayedAt"`
}

func printBeatmaps(beatmaps []Beatmap) {
	for _, beatmap := range beatmaps {
		printBeatmap(beatmap)
	}
}

func printBeatmap(b Beatmap) {
	values := reflect.ValueOf(b)
	// types := values.Type()
	for i := 0; i < values.NumField(); i++ {
		// fmt.Printf("%v: %v", types.Field(i).Name, values.Field(i))
		fmt.Printf("%v", values.Field(i))
		fmt.Println()
	}
	fmt.Println()
}

func parseDifficulty(s string) (string, error) {
	colorToDifficulty := map[string]string{
		"#00ced1": "easy",
		"#98fb98": "basic",
		"#ffa500": "advanced",
		"#fa8080": "expert",
		"#ee82ee": "master",
		"#ffceff": "re:master",
		"#ff5296": "utage",
	}

	re := regexp.MustCompile(`background-color:(#[0-9a-f]+)`)
	match := re.FindStringSubmatch(s)
	if len(match) < 2 {
		return "", fmt.Errorf("failed to parse difficulty")
	}
	color := match[1]
	return colorToDifficulty[color], nil
}

func getBeatmapsFromDB() ([]Beatmap, error) {
	dbURL := "http://localhost:8080/v1/beatmaps"
	req, _ := http.NewRequest("GET", dbURL, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return []Beatmap{}, err
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return []Beatmap{}, err
	}

	if r.StatusCode != http.StatusOK {
		bodyString := string(bodyBytes)
		return []Beatmap{}, fmt.Errorf("%v", bodyString)
	}

	beatmaps := []Beatmap{}
	json.Unmarshal(bodyBytes, &beatmaps)
	return beatmaps, nil
}

func dumpBeatmapsAsJson(beatmaps []Beatmap) error {
	jsonByte, err := json.Marshal(beatmaps)
	if err != nil {
		return err
	}

	// write to file
	os.MkdirAll("./tmp/json/", os.ModePerm)
	err = os.WriteFile("./tmp/json/beatmaps.json", jsonByte, 0666)
	if err != nil {
		return err
	}
	return nil
}

// if the beatmap has Touch notes -> dx beatmap
func determineBeatmapType(headerText string) string {
	if strings.Contains(headerText, "Touch") {
		headerText = "dx"
	}
	return "std"
}
