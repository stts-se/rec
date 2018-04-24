package rec

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
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
	Data     string `json:"data,omitempty"`
}

type ProcessInput struct {
	UserName    string             `json:"username"`
	Audio       Audio              `json:"audio"`
	Text        string             `json:"text"`
	RecordingID string             `json:"recording_id"`
	Weights     map[string]float64 `json:"weights,omitempty"`
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
	Ok                bool                 `json:"ok"`
	Confidence        float64              `json:"confidence"` // value between 0 and 1
	RecognitionResult string               `json:"recognition_result"`
	RecordingID       string               `json:"recording_id"`
	Message           string               `json:"message"`
	ComponentResults  []RecogniserResponse `json:"component_results,omitempty"`
}

type RecogniserResponse struct {
	Status            bool               `json:"status"`
	InputConfidence   map[string]float64 `json:"input_confidence,omitempty"` // recogniser, config, user, product
	Confidence        float64            `json:"confidence"`                 // value between 0 and 1
	RecognitionResult string             `json:"recognition_result"`
	RecordingID       string             `json:"recording_id"`
	Message           string             `json:"message"`
	Source            string             `json:"source"`
}

var spaceAndAfter = regexp.MustCompile(" .*$")

var prInputConfidenceRe = regexp.MustCompile("(\"input_confidence\": {)\n\\s*")
var prInputConfidenceChildrenRe = regexp.MustCompile("(\"(?:config|product|recogniser|user)\": [0-9.]+,?)\n\\s*(}?)")

func (pr ProcessResponse) PrettyJSONForced() string {
	res, _ := pr.PrettyJSON()
	return res
}

func (pr ProcessResponse) PrettyJSON() (string, error) {
	js, err := PrettyMarshal(pr)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response : %v", err)
	}
	res := string(js)
	res = prInputConfidenceRe.ReplaceAllString(res, "$1")
	res = prInputConfidenceChildrenRe.ReplaceAllString(res, "$1$2 ")
	res = strings.Replace(res, "} ,", "},", -1)
	return res, nil
}

func (pr ProcessResponse) Source() string {
	return spaceAndAfter.ReplaceAllString(pr.Message, "")
}

func (pr RecogniserResponse) String() string {
	status := "OK"
	if !pr.Status {
		status = "FAIL"
	}
	return fmt.Sprintf("[%s] %s |  %s %f %v %s %s", pr.Source, pr.RecognitionResult, status, pr.Confidence, pr.InputConfidence, pr.RecordingID, pr.Message)
}

func (pr ProcessResponse) String() string {
	status := "OK"
	if !pr.Ok {
		status = "FAIL"
	}
	return fmt.Sprintf("[%s] %s |  %s %v %s", pr.Message, pr.RecognitionResult, status, pr.Confidence, pr.RecordingID)
}
