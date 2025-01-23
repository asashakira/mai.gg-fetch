package user

type User struct {
	UserID   string `json:"userID,omitempty"`
	SegaID   string `json:"segaID,omitempty"`
	Password string `json:"password,omitempty"`
	GameName string `json:"gameName,omitempty"`
	TagLine  string `json:"tagLine,omitempty"`
}

type UserScrapeMetadata struct {
	UserID       string `json:"userID,omitempty"`
	LastPlayedAt string `json:"lastPlayedAt,omitempty"`
}
