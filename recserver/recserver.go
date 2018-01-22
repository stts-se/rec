package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const docTemplate = `
<!DOCTYPE html>
<html>
	<head>
               <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<title>{{.Title}}</title>
	</head>
	<body>
		{{range .Items}}<p><div>{{ .Desc }}</div><pre>{{ .Example }}</pre></p>{{else}}<div><strong>no rows</strong></div>{{end}}
	</body>
</html>`

type item struct {
	Desc    string
	Example string
}
type tplData struct {
	Title string
	Items []item
}

func prettyMarshal(thing interface{}) ([]byte, error) {
	var res []byte

	j, err := json.Marshal(thing)
	if err != nil {
		return res, err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, j, "", "\t")
	if err != nil {
		return res, err
	}
	res = prettyJSON.Bytes()
	return res, nil
}

func generateDoc(w http.ResponseWriter, r *http.Request) {
	title := "rec doc"
	processIn := processInput{
		UserName: "string",
		Audio: audio{
			FileType: "string",
			Data:     "string of base64 encoded data",
		},
		Text:        "string",
		RecordingID: "string",
	}

	processInSample := processInput{
		UserName: "user0001",
		Audio: audio{
			FileType: "audio/webm",
			Data:     "GkXfo59ChoEBQ ...",
		},
		Text:        "text to be spoken",
		RecordingID: "utterance_0001",
	}

	prettyJSON, err := prettyMarshal(processIn)
	if err != nil {
		msg := fmt.Sprintf("failed to pretty marshal : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	prettySampleJSON, err := prettyMarshal(processInSample)
	if err != nil {
		msg := fmt.Sprintf("failed to pretty marshal : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// processResp := processResponse{
	// 	Ok:                true,
	// 	Confidence:        0.0,
	// 	RecognitionResult: "string",
	// 	RecordingID:       "string",
	// 	Message:           "string",
	// }
	//prettyResponse, err := :prettyMarshal(processResp)

	t, err := template.New("webpage").Parse(docTemplate)
	if err != nil {
		msg := fmt.Sprintf("failed to parse doc template : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	s1 := string(prettyJSON)
	s2 := string(prettySampleJSON)

	//s3 := string(prettyResp)
	//_ = s3

	//log.Println(s1)
	//log.Println(s2)
	d := tplData{
		Title: title,
		Items: []item{
			item{Desc: "/rec/process/ input JSON to POST request", Example: s1},
			item{Desc: "/rec/process/ sample JSON input", Example: s2},
			item{Desc: "", Example: "__________________________________________________"},
		},
	}

	t.Execute(w, d)
}

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

	var ext string
	for _, e := range []string{"webm", "wav", "ogg", "mp3"} {
		if strings.Contains(rec.Audio.FileType, e) {
			ext = e
			break
		}
	}

	outFile := rec.RecordingID // filePath.Join(dirPath, rec.RecordingID + ". " + ext) "/tmp/nilz"
	if ext != "" {
		outFile = outFile + "." + ext
	} else {
		if rec.Audio.FileType == "" {
			log.Print("INPUT AUDIO FILE HAS NO ASSOCIATED FILE TYPE")
			// TODO Return error?
		} else {
			log.Printf("INPUT AUDIO FILE HAS UNKNOWN FILE TYPE '%s'\n", rec.Audio.FileType)
			// TODO Return error?
		}
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

func main() {

	audioDir = "audio_dir"
	_, sErr := os.Stat(audioDir)
	if os.IsNotExist(sErr) {
		os.Mkdir(audioDir, os.ModePerm)
	}

	p := "9993"
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.PathPrefix("/rec/recclient/").Handler(http.StripPrefix("/rec/recclient/", http.FileServer(http.Dir("../recclient"))))
	r.HandleFunc("/rec/", index)
	r.HandleFunc("/rec/process/", process).Methods("POST")
	r.HandleFunc("/rec/doc/", generateDoc).Methods("POST", "GET")

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
