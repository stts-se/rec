package main

import (
	"fmt"
	"os"

	"github.com/stts-se/rec"
)

type fileInfo struct {
	TargetUtterance   string  `json:"target_utterance"`
	ActualUtterance   string  `json:"actual_utterance"`
	Status            string  `json:"status"`
	Ok                bool    `json:"ok"`
	Confidence        float64 `json:"confidence"`
	RecognitionResult string  `json:"recognition_result"`
	Message           string  `json:"message"`
}

func writeJSONInfoFile(audioRef rec.AudioRef, rec rec.ProcessInput, res0 rec.ProcessResponse) error {

	infos := []fileInfo{}

	// writeMutex declaren in recserver.go
	writeMutex.Lock()
	defer writeMutex.Unlock()

	// Add matching info JSON file
	infoFilePath := audioRef.Path(".json")
	if _, err := os.Stat(infoFilePath); !os.IsNotExist(err) {
		os.Remove(infoFilePath)
	} // TODO Check for other err

	//fmt.Printf("writeJSONInfoFile DEBUG : %#v", res0)

	info := fileInfo{
		TargetUtterance:   rec.Text,
		Status:            "recogniser",
		Confidence:        res0.Confidence,
		RecognitionResult: res0.RecognitionResult,
		Message:           res0.Message,
	}
	infos = append(infos, info)
	for _, res := range res0.ComponentResults {

		info := fileInfo{
			TargetUtterance:   rec.Text,
			Status:            "recogniser",
			Confidence:        res.Confidence,
			RecognitionResult: res.RecognitionResult,
			Message:           res.Message,
		}
		infos = append(infos, info)
	}
	infoJSON, err := prettyMarshal(infos)
	if err != nil {
		return fmt.Errorf("writeJSONInfoFile: failed to create info JSON : %v", err)
	}
	infoFile, err := os.Create(infoFilePath)
	if err != nil {
		return fmt.Errorf("writeJSONInfoFile: failed to create info file : %v", err)
	}
	defer infoFile.Close()

	_, err = infoFile.WriteString(string(infoJSON) + "\n")
	if err != nil {
		return fmt.Errorf("writeJSONInfoFile: failed to write info file : %v", err)
	}

	return nil
}

// func writeJSONInfoFile(audioRef rec.AudioRef, rec rec.ProcessInput, res rec.ProcessResponse) error {

// 	// writeMutex declaren in recserver.go
// 	writeMutex.Lock()
// 	defer writeMutex.Unlock()

// 	// Add matching info JSON file

// 	//infoFileName := audioRef.BaseName /*rec.RecordingID*/ + ".json"
// 	//infoFilePath := filepath.Join(audioRef.BaseDir, audioRef.UserDir, infoFileName)
// 	infoFilePath := audioRef.Path(".json")
// 	if _, err := os.Stat(infoFilePath); !os.IsNotExist(err) {
// 		os.Remove(infoFilePath)
// 	} // TODO Check for other err

// 	info := fileInfo{
// 		TargetUtterance:   rec.Text,
// 		Status:            "recogniser",
// 		RecognitionResult: res.RecognitionResult,
// 	}

// 	infoJSON, err := prettyMarshal(info)
// 	if err != nil {
// 		return fmt.Errorf("writeJSONInfoFile: failed to create info JSON : %v", err)
// 	}
// 	infoFile, err := os.Create(infoFilePath)
// 	if err != nil {
// 		return fmt.Errorf("writeJSONInfoFile: failed to create info file : %v", err)
// 	}
// 	defer infoFile.Close()

// 	_, err = infoFile.WriteString(string(infoJSON) + "\n")
// 	if err != nil {
// 		return fmt.Errorf("writeJSONInfoFile: failed to write info file : %v", err)
// 	}

// 	return nil
// }
