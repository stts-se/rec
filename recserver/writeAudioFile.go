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

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/audioproc"
)

var noiseRedSuffix = "-noisered"

func validAudioFileExtension(ext string) bool {
	return (ext == "opus" || ext == "mp3" || ext == "wav")
}

func writeAudioFile(audioDir string, rec rec.ProcessInput) (string, error) {
	if strings.TrimSpace(audioDir) == "" {
		return "", fmt.Errorf("writeAudioFile: empty audioDir argument")
	}
	if strings.TrimSpace(rec.UserName) == "" {
		return "", fmt.Errorf("writeAudioFile: empty input username")
	}

	dirPath := filepath.Join(audioDir, rec.UserName)
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		// First file to save for rec.Username, create dir of
		// user name
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("writeAudioFile: failed to create dir : %v", err)
		}
	}

	if strings.TrimSpace(rec.Audio.FileType) == "" {
		msg := fmt.Sprintf("input audio for '%s' has no associated file type", rec.RecordingID)
		log.Print(msg)
		return "", fmt.Errorf(msg)
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
		return "", fmt.Errorf(msg)
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
		return "", fmt.Errorf("%s : %v", msg, err)
	}

	err = ioutil.WriteFile(audioFilePath, audio, 0644)
	if err != nil {
		msg := fmt.Sprintf("failed to write audio file : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		//http.Error(w, msg, http.StatusBadRequest)
		return "", fmt.Errorf("%s : %v", msg, err)
	}
	log.Printf("AUDIO LEN: %d\n", len(audio))
	log.Printf("WROTE FILE: %s\n", audioFile)

	// Convert to wav, while we're at it:
	if ext != "wav" {
		// ffmpegConvert function is defined in ffmpegConvert.go
		audioFilePathWav := filepath.Join(dirPath, rec.RecordingID+".wav")
		audioFilePathWavReduced := filepath.Join(dirPath, rec.RecordingID+noiseRedSuffix+".wav")
		err = ffmpegConvert(audioFilePath, audioFilePathWav, false)
		if err != nil {
			msg := fmt.Sprintf("writeAudioFile failed converting from %s to %s : %v", audioFilePath, audioFilePathWav, err)
			log.Print(msg)
			return "", fmt.Errorf(msg)
		}
		if audioproc.SoxEnabled() {
			err = audioproc.NoiseReduce(audioFilePathWav, audioFilePathWavReduced)
			if err != nil {
				msg := fmt.Sprintf("writeAudioFile failed noise reduction for file : %v", err)
				log.Print(msg)
				return "", fmt.Errorf(msg)
			}
			log.Printf("Converted saved file into noise-reduced wav: %s", audioFilePathWavReduced)
		} else { // silently skip generation of wav with noise reduction and remove old noisered file if it exists
			if _, err = os.Stat(audioFilePathWavReduced); !os.IsNotExist(err) {
				err = os.Remove(audioFilePathWavReduced)
				if err != nil {
					log.Printf("failed to remove file : %v\n", err)
				}
			}
		}
		log.Printf("Converted saved file into wav: %s", audioFilePathWav)
	}

	// Convert to opus, while we're at it:
	if defaultExtension == "opus" {
		if ext != defaultExtension {
			audioFilePathOpus := filepath.Join(dirPath, rec.RecordingID+".opus")
			// ffmpegConvert function is defined in ffmpegConvert.go
			err = ffmpegConvert(audioFilePath, audioFilePathOpus, false)
			if err != nil {
				msg := fmt.Sprintf("writeAudioFile failed converting from %s to %s : %v", audioFilePath, audioFilePathOpus, err)
				log.Print(msg)
				return "", fmt.Errorf(msg)
			} // Woohoo, file converted into opus
			log.Printf("Converted saved file into opus: %s", audioFilePathOpus)
		}
	} else if defaultExtension == "mp3" {
		if ext != defaultExtension+".mp3" {
			audioFilePathMp3 := filepath.Join(dirPath, rec.RecordingID+".mp3")
			// ffmpegConvert function is defined in ffmpegConvert.go
			err = ffmpegConvert(audioFilePath, audioFilePathMp3, false)
			if err != nil {
				msg := fmt.Sprintf("writeAudioFile failed converting from %s to %s : %v", audioFilePath, audioFilePathMp3, err)
				log.Print(msg)
				return "", fmt.Errorf(msg)
			} // Woohoo, file converted into mp3
			log.Printf("Converted saved file into mp3: %s", audioFilePathMp3)
		}
	}

	if !validAudioFileExtension(defaultExtension) {
		msg := fmt.Sprintf("writeAudioFile unknown default extension: %s", defaultExtension)
		log.Print(msg)
		return "", fmt.Errorf(msg)
	}

	//err = writeJSONInfoFile(dirPath, rec)

	audioFileFinal := filepath.Join(dirPath, rec.RecordingID+"."+defaultExtension)
	return audioFileFinal, nil
}
