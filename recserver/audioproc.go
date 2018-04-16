package main

import (
	// "encoding/base64"
	// "encoding/json"
	// "fmt"
	// "github.com/gorilla/mux"
	// "io/ioutil"
	"log"
	// "net/http"
	// "os"
	"os/exec"
	// "path/filepath"
	"regexp"

	//"github.com/stts-se/rec/audioproc"
	"github.com/stts-se/rec/config"
)

func soxEnabled() bool {
	soxCmd := config.MyConfig.SoxCommand
	_, pErr := exec.LookPath(soxCmd)
	if pErr != nil {
		log.Printf("recserver.FfmpegEnabled(): External '%s' command does not exist!", soxCmd)
		return false
	}
	return true
}

var onRegexp = regexp.MustCompile("^(?i)(true|yes|y|1|on)$")

// func buildSpectrogram(w http.ResponseWriter, r *http.Request) {
// 	var res audioResponse
// 	vars := mux.Vars(r)
// 	userName := vars["username"]
// 	utteranceID := vars["utterance_id"]
// 	// TODO Remove noise reduced variants?
// 	noiseRedS := getParam("noise_red", r)
// 	var ext = vars["ext"]
// 	if ext == "" {
// 		ext = defaultExtension
// 	}

// 	useNoiseReduction := false
// 	if onRegexp.MatchString(noiseRedS) {
// 		useNoiseReduction = true
// 	}

// 	_, err := os.Stat(filepath.Join(audioDir, userName))
// 	if os.IsNotExist(err) {
// 		msg := fmt.Sprintf("get_spectrogram: no audio dir for user '%s'", userName)
// 		log.Print(msg)
// 		http.Error(w, msg, http.StatusBadRequest)
// 		return
// 	}

// 	audioFile := filepath.Join(audioDir, userName, utteranceID+"."+ext)
// 	specFile := filepath.Join(audioDir, userName, utteranceID+".png")
// 	_, err = os.Stat(audioFile)
// 	if os.IsNotExist(err) {
// 		msg := fmt.Sprintf("get_spectrogram: no audio for utterance '%s'", utteranceID)
// 		log.Print(msg)
// 		http.Error(w, msg, http.StatusBadRequest)
// 		return
// 	}

// 	err = audioproc.BuildSoxSpectrogram(audioFile, specFile, useNoiseReduction)
// 	if err != nil {
// 		msg := fmt.Sprintf("get_spectrogram: failed to build spectrogram : %v", err)
// 		log.Print(msg)
// 		http.Error(w, msg, http.StatusInternalServerError)
// 		return
// 	}

// 	bytes, err := ioutil.ReadFile(specFile)
// 	if err != nil {
// 		msg := fmt.Sprintf("get_spectrogram: failed to read spectrogram file : %v", err)
// 		log.Print(msg)
// 		http.Error(w, msg, http.StatusInternalServerError)
// 		return
// 	}

// 	data := base64.StdEncoding.EncodeToString(bytes)

// 	res.FileType = "image/png"
// 	res.Data = data

// 	resJSON, err := prettyMarshal(res)
// 	if err != nil {
// 		msg := fmt.Sprintf("get_spectrogram: failed to create JSON from struct : %v", res)
// 		log.Print(msg)
// 		http.Error(w, msg, http.StatusBadRequest)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	fmt.Fprintf(w, "%s\n", string(resJSON))
// }
