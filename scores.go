package main

type Score struct {
	ScoreID   string `json:"scoreID"`
	BeatmapID string `json:"beatmapID"`
	SongID    string `json:"songID"`
	UserID    string `json:"userID"`
	Accuracy  string `json:"accuracy"`
	MaxCombo  int32  `json:"maxCombo"`
	DxScore   int32  `json:"dxScore"`
	PlayedAt  string `json:"playedAt"`
}
