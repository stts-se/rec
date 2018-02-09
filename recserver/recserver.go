package main

// TODO put a mutex around file reading and writing

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
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

	//HB testing
	res, err = runExternalKaldiDecoder("audio_dir/user0001/rec_0001.wav", res)
	if err != nil {
		msg := fmt.Sprintf("failed decoding audio file : %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// TODO This is weird. Structs 'processInput' and
	// 'processRespons' and 'infoFile' should probably be a single
	// struct

	// writeJSONInfoFile defined in writeJSONInfoFile.go
	err = writeJSONInfoFile(audioDir, input, res)
	if err != nil {
		msg := fmt.Sprintf("failed writing info file : %v", err)
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

		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(resJSON))
}

type audioResponse struct {
	FileType string `json:"file_type"`
	Data     string `json:"data"`
	Message  string `json:"message"`
}

func getAudio(w http.ResponseWriter, r *http.Request) {
	var res audioResponse
	vars := mux.Vars(r)
	userName := vars["username"]
	utteranceID := vars["utterance_id"]
	_, err := os.Stat(filepath.Join(audioDir, userName))
	if os.IsNotExist(err) {
		msg := fmt.Sprintf("get_audio: no audio for user '%s'", userName)
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	audioFile := filepath.Join(audioDir, userName, utteranceID+".wav")
	_, err = os.Stat(audioFile)
	if os.IsNotExist(err) {
		msg := fmt.Sprintf("get_audio: no audio for utterance '%s'", utteranceID)
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	bytes, err := ioutil.ReadFile(audioFile)
	if err != nil {
		msg := fmt.Sprintf("get_audio: failed to read audio file : %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	data := base64.StdEncoding.EncodeToString(bytes)

	res.FileType = "audio/wav"
	res.Data = data

	resJSON, err := prettyMarshal(res)
	if err != nil {
		msg := fmt.Sprintf("get_audio: failed to create JSON from struct : %v", res)
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(resJSON))
}

// The path to the directory where audio files are saved
var audioDir string

// This is filled in by main, listing the URLs handled by the router,
// so that these can be shown in the generated docs.
var walkedURLs []string

func main() {

	// TODO return text prompt etc as well

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

	//func loadUtteranceLists defined in getUtterance.go
	uls, err := loadUtteranceLists(audioDir)
	if err != nil {
		msg := fmt.Sprintf("failed to load user utterance lists : %v", err)
		log.Print(msg)
		os.Exit(1)
	}
	uttLists = uls

	p := "9993"
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/rec/", index)
	r.HandleFunc("/rec/process/", process).Methods("POST")
	// generateDoc is definied in file generateDoc.go
	r.HandleFunc("/rec/doc/", generateDoc).Methods("GET")

	// TODO Should this rather be a POST request?
	r.HandleFunc("/rec/get_audio/{username}/{utterance_id}", getAudio).Methods("GET")

	// Defined in getUtterance.go
	r.HandleFunc("/rec/get_next_utterance/{username}", getNextUtterance).Methods("GET")
	r.HandleFunc("/rec/get_previous_utterance/{username}", getPreviousUtterance).Methods("GET")

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
