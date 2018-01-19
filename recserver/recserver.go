package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	ErrorMessage      string  `json:"error_message"`
}

func process(w http.ResponseWriter, r *http.Request) {
	res := processResponse{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read request body : %v", err)
		res.ErrorMessage = msg
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	input := processInput{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal incoming JSON : %v", err)
		res.ErrorMessage = msg
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	log.Printf("GOT username: %s\ttext: %s\t recording id: %s\n", input.UserName, input.Text, input.RecordingID)

	var audio []byte
	audio, err = base64.StdEncoding.DecodeString(input.Audio.Data)
	if err != nil {
		msg := fmt.Sprintf("failed audio base64 decoding : %v", err)
		res.ErrorMessage = msg
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var ext string
	for _, e := range []string{"webm", "wav", "ogg", "mp3"} {
		if strings.Contains(input.Audio.FileType, e) {
			ext = e
			break
		}
	}

	outFile := "/tmp/nilz"
	if ext != "" {
		outFile = outFile + "." + ext
	} else {
		if input.Audio.FileType == "" {
			log.Print("INPUT AUDIO FILE HAS NO ASSOCIATED FILE TYPE")
		} else {
			log.Printf("INPUT AUDIO FILE HAS UNKNOWN FILE TYPE '%s'\n", input.Audio.FileType)
		}
	}

	if _, err = os.Stat(outFile); !os.IsNotExist(err) {
		os.Remove(outFile)
	}

	err = ioutil.WriteFile(outFile, audio, 0644)
	if err != nil {
		msg := fmt.Sprintf("failed to write audio file : %v", err)
		res.ErrorMessage = msg
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	log.Printf("AUDIO LEN: %d\n", len(audio))
	log.Printf("WROTE FILE: %s\n", outFile)

	// TODO Create reasonable response

	res.Ok = true
	res.RecordingID = input.RecordingID

	resJSON, err := json.Marshal(res)
	if err != nil {
		msg := fmt.Sprintf("failed to marshal response : %v", err)
		res.ErrorMessage = msg
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(resJSON))
}
func main() {
	p := "9993"
	r := mux.NewRouter()
	r.StrictSlash(false)
	r.PathPrefix("/recclient/").Handler(http.StripPrefix("/recclient/", http.FileServer(http.Dir("../recclient"))))
	r.HandleFunc("/", index)
	r.HandleFunc("/process/", process).Methods("POST", "GET")
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
