package main

import (
	"encoding/base64"
	//"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func writeAudioFile(audioDir string, rec processInput) error {
	if strings.TrimSpace(audioDir) == "" {
		return fmt.Errorf("writeAudioFile: empty audioDir argument")
	}
	if strings.TrimSpace(rec.UserName) == "" {
		return fmt.Errorf("writeAudioFile: empty input username")
	}

	dirPath := filepath.Join(audioDir, rec.UserName)
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		// First file to save for rec.Username, create dir of
		// user name
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("writeAudioFile: failed to create dir : %v", err)
		}
	}

	if strings.TrimSpace(rec.Audio.FileType) == "" {
		msg := fmt.Sprintf("input audio for '%s' has no associated file type", rec.RecordingID)
		log.Print(msg)
		return fmt.Errorf(msg)
	}

	var ext string
	for _, e := range []string{"webm", "wav", "ogg", "mp3"} {
		if strings.Contains(rec.Audio.FileType, e) {
			ext = e
			break
		}
	}

	if ext == "" {
		msg := fmt.Sprintf("unknown file type for '%s': %s", rec.RecordingID, rec.Audio.FileType)
		log.Print(msg)
		return fmt.Errorf(msg)
	}

	audioFile := rec.RecordingID // filePath.Join(dirPath, rec.RecordingID + ". " + ext) "/tmp/nilz"
	if ext != "" {
		audioFile = audioFile + "." + ext
	}

	audioFilePath := filepath.Join(dirPath, audioFile)
	// If file of same name exists, remove
	if _, err = os.Stat(audioFilePath); !os.IsNotExist(err) {
		os.Remove(audioFilePath)
	}

	var audio []byte
	audio, err = base64.StdEncoding.DecodeString(rec.Audio.Data)
	if err != nil {
		msg := fmt.Sprintf("failed audio base64 decoding : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		//http.Error(w, msg, http.StatusBadRequest)
		return fmt.Errorf("%s : %v", msg, err)
	}

	err = ioutil.WriteFile(audioFilePath, audio, 0644)
	if err != nil {
		msg := fmt.Sprintf("failed to write audio file : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		//http.Error(w, msg, http.StatusBadRequest)
		return fmt.Errorf("%s : %v", msg, err)
	}
	log.Printf("AUDIO LEN: %d\n", len(audio))
	log.Printf("WROTE FILE: %s\n", audioFile)

	// Conver to wav, while we're at it:
	if ext != "wav" {
		audioFilePathWav := filepath.Join(dirPath, rec.RecordingID+".wav")
		// ffmpegConvert function is defined in ffmpegConvert.go
		err = ffmpegConvert(audioFilePath, audioFilePathWav, false)
		if err != nil {
			msg := fmt.Sprintf("writeAudioFile failed converting file : %v", err)
			log.Print(msg)
			return fmt.Errorf(msg)
		} // Woohoo, file converted into wav
		log.Printf("Converted saved file into wav: %s", audioFilePathWav)
	}

	//err = writeJSONInfoFile(dirPath, rec)

	return nil
}
