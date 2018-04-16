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
	//"github.com/stts-se/rec/audioproc"
)

// TODO Remove noise reduced variants?
//var noiseRedSuffix = "-noisered"

func validAudioFileExtension(ext string) bool {
	return ext == "wav"
	//return (ext == "opus" || ext == "mp3" || ext == "wav")
}

// func noiseReduce(audioFilePathWav string, audioRef rec.AudioRef) error {
// 	audioFilePathWavReduced := audioRef.Path(noiseRedSuffix + ".wav")
// 	err := audioproc.NoiseReduce(audioFilePathWav, audioFilePathWavReduced)
// 	if err != nil {
// 		msg := fmt.Sprintf("writeAudioFile failed noise reduction for file : %v", err)
// 		log.Print(msg)
// 		return fmt.Errorf(msg)
// 	}
// 	log.Printf("Converted saved file into noise-reduced wav: %s", audioFilePathWavReduced)
// 	return nil
// }

// save the original audio from the client + a set of additional versions
func writeAudioFile(audioDir rec.AudioDir, input rec.ProcessInput) (rec.AudioFile, error) {

	// writeMutex declaren in recserver.go
	writeMutex.Lock()
	defer writeMutex.Unlock()

	if strings.TrimSpace(audioDir.BaseDir) == "" {
		return rec.AudioFile{}, fmt.Errorf("writeAudioFile: empty audioDir argument")
	}
	if strings.TrimSpace(audioDir.UserDir) == "" {
		return rec.AudioFile{}, fmt.Errorf("writeAudioFile: empty userDir argument")
	}
	if strings.TrimSpace(input.UserName) == "" {
		return rec.AudioFile{}, fmt.Errorf("writeAudioFile: empty input username")
	}

	userDir := audioDir.Path()
	inputAudioDirPath := filepath.Join(userDir, inputAudioDir)

	_, err := os.Stat(userDir)
	if os.IsNotExist(err) {
		// First file to save for input.Username, create dir of
		// user name
		err = os.MkdirAll(userDir, os.ModePerm)
		if err != nil {
			return rec.AudioFile{}, fmt.Errorf("writeAudioFile: failed to create dir : %v", err)
		}
	}
	_, err = os.Stat(inputAudioDirPath)
	if os.IsNotExist(err) {
		// create subdir input_audio to keep original audio from client
		err = os.MkdirAll(inputAudioDirPath, os.ModePerm)
		if err != nil {
			return rec.AudioFile{}, fmt.Errorf("writeAudioFile: failed to create dir : %v", err)
		}
	}

	if strings.TrimSpace(input.Audio.FileType) == "" {
		msg := fmt.Sprintf("input audio for '%s' has no associated file type", input.RecordingID)
		log.Print(msg)
		return rec.AudioFile{}, fmt.Errorf(msg)
	}

	var ext string
	for _, e := range []string{"webm", "wav", "ogg", "mp3"} {
		if strings.Contains(input.Audio.FileType, e) {
			ext = e
			break
		}
	}

	if ext == "" {
		msg := fmt.Sprintf("unknown file type for '%s': %s", input.RecordingID, input.Audio.FileType)
		log.Print(msg)
		return rec.AudioFile{}, fmt.Errorf(msg)
	}

	// generate next running number for file with same recordingID. Starts at "0001"
	// always returns, with default returnvaule "0001"
	// declared in generateNextFileNum.go
	runningNum := generateNextFileNum(audioDir, input.RecordingID)
	audioRef := rec.AudioRef{Dir: audioDir, BaseName: input.RecordingID + "_" + runningNum}
	//audioFilePath := audioRef.Path("." + ext)

	// If file of same name exists, remove
	//if _, err = os.Stat(audioFilePath); !os.IsNotExist(err) {
	//	os.Remove(audioFilePath)
	//}

	var audio []byte
	audio, err = base64.StdEncoding.DecodeString(input.Audio.Data)
	if err != nil {
		msg := fmt.Sprintf("failed audio base64 decoding : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		//http.Error(w, msg, http.StatusBadRequest)
		return rec.AudioFile{}, fmt.Errorf("%s : %v", msg, err)
	}

	inputAudioFilePath := filepath.Join(inputAudioDirPath, audioRef.FileName("."+ext))

	// (1) Save original audio input file (whatever extension/format)
	err = ioutil.WriteFile(inputAudioFilePath, audio, 0644)
	if err != nil {
		msg := fmt.Sprintf("failed to write audio file : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		//http.Error(w, msg, http.StatusBadRequest)
		return rec.AudioFile{}, fmt.Errorf("%s : %v", msg, err)
	}
	//log.Printf("AUDIO LEN: %d\n", len(audio))
	log.Printf("WROTE FILE: %s\n", inputAudioFilePath)

	// (2) ALWAYS convert to wav 16kHz MONO
	// ffmpegConvert function is defined in ffmpegConvert.go
	audioFilePathWav := audioRef.Path(".wav")
	err = ffmpegConvert(inputAudioFilePath, audioFilePathWav, false)
	if err != nil {
		msg := fmt.Sprintf("writeAudioFile failed converting from %s to %s : %v", inputAudioFilePath, audioFilePathWav, err)
		log.Print(msg)
		return rec.AudioFile{}, fmt.Errorf(msg)
	}
	log.Printf("Converted saved file into wav: %s", audioFilePathWav)

	// err = noiseReduce(audioFilePathWav, audioRef)
	// if err != nil {
	// 	return rec.AudioFile{}, err
	// }

	// if defaultExtension == "opus" {
	// 	if ext != defaultExtension {
	// 		audioFilePathOpus := audioRef.Path(".opus")
	// 		err = ffmpegConvert(inputAudioFilePath, audioFilePathOpus, false)
	// 		if err != nil {
	// 			msg := fmt.Sprintf("writeAudioFile failed converting from %s to %s : %v", inputAudioFilePath, audioFilePathOpus, err)
	// 			log.Print(msg)
	// 			return rec.AudioFile{}, fmt.Errorf(msg)
	// 		}
	// 		log.Printf("Converted saved file into opus: %s", audioFilePathOpus)
	// 	}
	// } else if defaultExtension == "mp3" {
	// 	if ext != defaultExtension+".mp3" {
	// 		audioFilePathMp3 := audioRef.Path(".mp3")
	// 		err = ffmpegConvert(inputAudioFilePath, audioFilePathMp3, false)
	// 		if err != nil {
	// 			msg := fmt.Sprintf("writeAudioFile failed converting from %s to %s : %v", inputAudioFilePath, audioFilePathMp3, err)
	// 			log.Print(msg)
	// 			return rec.AudioFile{}, fmt.Errorf(msg)
	// 		}
	// 		log.Printf("Converted saved file into mp3: %s", audioFilePathMp3)
	// 	}
	// }

	if !validAudioFileExtension(defaultExtension) {
		msg := fmt.Sprintf("writeAudioFile unknown default extension: %s", defaultExtension)
		log.Print(msg)
		return rec.AudioFile{}, fmt.Errorf(msg)
	}

	//err = writeJSONInfoFile(dirPath, rec)

	audioFileFinal := rec.AudioFile{
		BasePath:  audioRef,
		Extension: ("." + defaultExtension),
	}
	log.Printf("writeAudioFile: audioFileFinal=%v\n", audioFileFinal)
	return audioFileFinal, nil
}
