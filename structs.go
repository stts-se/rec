package rec

type Utterance struct {
	UserName    string `json:"username"`
	Text        string `json:"text"`
	RecordingID string `json:"recording_id"`
	Message     string `json:"message"`
	Num         int    `json:"num"`
	Of          int    `json:"of"`
}

type UttList struct {
	Name string      `json:"name"`
	Utts []Utterance `json:"utts"`
}
