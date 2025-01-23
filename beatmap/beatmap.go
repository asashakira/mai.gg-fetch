package beatmap

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
)

type Beatmap struct {
	BeatmapID     string  `json:"beatmapID,omitempty"`
	SongID        string  `json:"songID,omitempty"`
	Difficulty    string  `json:"difficulty,omitempty"`
	Level         string  `json:"level,omitempty"`
	InternalLevel float64 `json:"internalLevel,omitempty"`
	Type          string  `json:"type,omitempty"`
	TotalNotes    int32   `json:"totalNotes,omitempty"`
	Tap           int32   `json:"tap,omitempty"`
	Hold          int32   `json:"hold,omitempty"`
	Slide         int32   `json:"slide,omitempty"`
	Touch         int32   `json:"touch,omitempty"`
	Break         int32   `json:"break,omitempty"`
	NoteDesigner  string  `json:"noteDesigner,omitempty"`
	MaxDxScore    int32   `json:"maxDxScore,omitempty"`
	IsValid       bool    `json:"isValid,omitempty"`
}

func PrintBeatmaps(beatmaps []Beatmap) {
	for _, beatmap := range beatmaps {
		PrintBeatmap(beatmap)
	}
}

func PrintBeatmap(b Beatmap) {
	values := reflect.ValueOf(b)
	// types := values.Type()
	for i := 0; i < values.NumField(); i++ {
		// fmt.Printf("%v: %v", types.Field(i).Name, values.Field(i))
		fmt.Printf("%v", values.Field(i))
		fmt.Println()
	}
	fmt.Println()
}

func ParseDifficulty(s string) (string, error) {
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

func DumpBeatmapsAsJson(beatmaps []Beatmap) error {
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
