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

type Audio struct {
	FileType string `json:"file_type"`
	Data     string `json:"data"`
}

type ProcessInput struct {
	UserName    string `json:"username"`
	Audio       Audio  `json:"audio"`
	Text        string `json:"text"`
	RecordingID string `json:"recording_id"`
}
