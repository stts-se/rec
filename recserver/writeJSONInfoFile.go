package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type fileInfo struct {
	TargetUtterance   string  `json:"target_utterance"`
	ActualUtterance   string  `json:"actual_utterance"`
	Status            string  `json:"status"`
	Ok                bool    `json:"ok"`
	Confidence        float32 `json:"confidence"`
	RecognitionResult string  `json:"recognition_result"`
}

func writeJSONInfoFile(audioDir string, rec processInput, res processResponse) error {

	// Add matching info JSON file

	infoFileName := rec.RecordingID + ".json"
	infoFilePath := filepath.Join(audioDir, rec.UserName, infoFileName)
	if _, err := os.Stat(infoFilePath); !os.IsNotExist(err) {
		os.Remove(infoFilePath)
	} // TODO Check for other err

	info := fileInfo{
		TargetUtterance:   rec.Text,
		Status:            "recogniser",
		RecognitionResult: res.RecognitionResult,
	}

	infoJSON, err := prettyMarshal(info)
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
