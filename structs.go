package rec

import (
	"fmt"
	"path/filepath"
)

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

type AudioDir struct {
	BaseDir string
	UserDir string
}

func (ad AudioDir) Path() string {
	return filepath.Join(ad.BaseDir, ad.UserDir)
}

type AudioRef struct {
	Dir      AudioDir
	BaseName string // without file extension
}

func (ar AudioRef) Path(extension string) string {
	fName := ar.fileName(extension)
	return filepath.Join(ar.Dir.Path(), fName)
}

func (ar AudioRef) fileName(extension string) string {
	return fmt.Sprintf("%s%s", ar.BaseName, extension)
}

type AudioFile struct {
	BasePath  AudioRef
	Extension string
}

func (af AudioFile) Path() string {
	fName := af.BasePath.fileName(af.Extension)
	return filepath.Join(af.BasePath.Dir.Path(), fName)
}

type ProcessResponse struct {
	Ok                bool    `json:"ok"`
	Confidence        float32 `json:"confidence"` // value between 0 and 1
	RecognitionResult string  `json:"recognition_result"`
	RecordingID       string  `json:"recording_id"`
	Message           string  `json:"message"`
}
