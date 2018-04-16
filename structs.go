package rec

import (
	"fmt"
	"path/filepath"
	"regexp"
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

func NewAudioRef(baseDir string, userDir string, baseName string) AudioRef {
	dir := AudioDir{BaseDir: baseDir, UserDir: userDir}
	return AudioRef{Dir: dir, BaseName: baseName}
}

func (ar AudioRef) Path(extension string) string {
	fName := ar.FileName(extension)
	return filepath.Join(ar.Dir.Path(), fName)
}

func (ar AudioRef) FileName(extension string) string {
	return fmt.Sprintf("%s%s", ar.BaseName, extension)
}

type AudioFile struct {
	BasePath  AudioRef
	Extension string
}

func NewAudioFile(baseDir string, userDir string, baseName string, extension string) AudioFile {
	audioRef := NewAudioRef(baseDir, userDir, baseName)
	return AudioFile{BasePath: audioRef, Extension: extension}
}

func (af AudioFile) Path() string {
	fName := af.BasePath.FileName(af.Extension)
	return filepath.Join(af.BasePath.Dir.Path(), fName)
}

func (af AudioFile) AudioDir() AudioDir {
	return af.BasePath.Dir
}

type ProcessResponse struct {
	Ok                bool    `json:"ok"`
	Confidence        float32 `json:"confidence"` // value between 0 and 1
	RecognitionResult string  `json:"recognition_result"`
	RecordingID       string  `json:"recording_id"`
	//Source            string  `json:"source"` // add later
	Message string `json:"message"`
}

var sourceFromMessageRe = regexp.MustCompile("^(\\[[^\\]]+\\]).*")

func (pr ProcessResponse) String() string {
	source := sourceFromMessageRe.ReplaceAllString(pr.Message, "$1")
	if source == "" || source == pr.Message {
		source = "[undef source]"
	}
	status := "OK  "
	if !pr.Ok {
		status = "FAIL"
	}
	return fmt.Sprintf("%-18s %s %s | %f %s %s", source, status, pr.RecognitionResult, pr.Confidence, pr.Message, pr.RecordingID)
}
