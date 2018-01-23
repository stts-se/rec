package main

import (
	//"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	//"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../recclient/index.html")
}

type audio struct {
	FileType string `json:"file_type"`
	Data     string `json:"data"`
}

// {username, audio, text, (recording_id if overwriting)}
type processInput struct {
	UserName    string `json:"username"`
	Audio       audio  `json:"audio"`
	Text        string `json:"text"`
	RecordingID string `json:"recording_id"`
}

//	{
//	 "ok": true|false,
//	 ? "confidence": <percent-value>,
//	 ? "recognition_result": <text-string>,
//	 ? "recording_id": <uri>
//	}
type processResponse struct {
	Ok                bool    `json:"ok"`
	Confidence        float32 `json:"confidence"`
	RecognitionResult string  `json:"recognition_result"`
	RecordingID       string  `json:"recording_id"`
	Message           string  `json:"message"`
}

func writeAudioFile(audioDir string, rec processInput) error {
	dirPath := filepath.Join(audioDir, rec.UserName)
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) { // First file to save for rec.Username
		os.MkdirAll(dirPath, os.ModePerm)
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

	outFile := rec.RecordingID // filePath.Join(dirPath, rec.RecordingID + ". " + ext) "/tmp/nilz"
	if ext != "" {
		outFile = outFile + "." + ext
	}

	outFilePath := filepath.Join(dirPath, outFile)
	if _, err = os.Stat(outFilePath); !os.IsNotExist(err) {
		os.Remove(outFilePath)
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

	err = ioutil.WriteFile(outFilePath, audio, 0644)
	if err != nil {
		msg := fmt.Sprintf("failed to write audio file : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		//http.Error(w, msg, http.StatusBadRequest)
		return fmt.Errorf("%s : %v", msg, err)
	}
	log.Printf("AUDIO LEN: %d\n", len(audio))
	log.Printf("WROTE FILE: %s\n", outFile)

	// Conver to wav, while we're at it:
	if ext != "wav" {
		outFilePathWav := filepath.Join(dirPath, rec.RecordingID+".wav")
		// ffmpegConvert function is defined in ffmpegConvert.go
		err = ffmpegConvert(outFilePath, outFilePathWav, false)
		if err != nil {
			msg := fmt.Sprintf("writeAudioFile failed converting file : %v", err)
			log.Print(msg)
			return fmt.Errorf(msg)
		} // Woohoo, file converted into wav
		log.Print("Converted saved file into wav: %v", outFilePathWav)
	}

	return nil
}

func checkProcessInput(input processInput) error {
	var errMsg []string

	if strings.TrimSpace(input.UserName) == "" {
		errMsg = append(errMsg, "no value for 'username'")
	}
	if strings.TrimSpace(input.Text) == "" {
		errMsg = append(errMsg, "no value for 'text'")
	}
	if strings.TrimSpace(input.RecordingID) == "" {
		errMsg = append(errMsg, "no value for 'recording_id'")
	}

	if len(input.Audio.Data) == 0 {
		errMsg = append(errMsg, "no 'audio.data'")
	}
	if strings.TrimSpace(input.Audio.FileType) == "" {
		errMsg = append(errMsg, "no value for 'audio.file_type'")
	}

	if len(errMsg) > 0 {
		return fmt.Errorf("missing values in input JSON: %s", strings.Join(errMsg, " : "))
	}

	return nil
}

func process(w http.ResponseWriter, r *http.Request) {
	res := processResponse{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read request body : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	input := processInput{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal incoming JSON : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	err = checkProcessInput(input)
	if err != nil {
		msg := fmt.Sprintf("incoming JSON was incomplete: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	log.Printf("GOT username: %s\ttext: %s\t recording id: %s\n", input.UserName, input.Text, input.RecordingID)

	err = writeAudioFile(audioDir, input)
	if err != nil {
		msg := fmt.Sprintf("failed writing audio file : %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	// TODO Create reasonable response

	res.Ok = true
	res.RecordingID = input.RecordingID

	resJSON, err := json.Marshal(res)
	if err != nil {
		msg := fmt.Sprintf("failed to marshal response : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(resJSON))
}

var audioDir string

// This is filled in by main, listing the URLs handled by the router
var walkedURLs []string

func main() {

	//TODO Check Go ffmpeg, or similar, bindings instead of
	// external call

	// Test that external ffmpeg command is found, or exit
	cmd := "ffmpeg"
	_, pErr := exec.LookPath(cmd)
	if pErr != nil {

		log.Printf("Exiting. Failed to find external command '%s'. Try installing it.\n", cmd)
		os.Exit(0)
	}

	audioDir = "audio_dir"
	_, sErr := os.Stat(audioDir)
	if os.IsNotExist(sErr) {
		os.Mkdir(audioDir, os.ModePerm)
	}

	p := "9993"
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/rec/", index)
	r.HandleFunc("/rec/process/", process).Methods("POST")
	// generateDoc is definied in file generateDoc.go
	r.HandleFunc("/rec/doc/", generateDoc).Methods("POST", "GET")

	// List route URLs to use as simple on-line documentation
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		walkedURLs = append(walkedURLs, t)
		return nil
	})

	// Add paths that don't need to be in the generated
	// documentation afte the r.Walk above

	r.PathPrefix("/rec/recclient/").Handler(http.StripPrefix("/rec/recclient/", http.FileServer(http.Dir("../recclient"))))

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:" + p,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("rec server started on port " + p)
	log.Fatal(srv.ListenAndServe())
	fmt.Println("No fun")
}
