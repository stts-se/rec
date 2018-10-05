package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/stts-se/rec"
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

// Generates simple documentation semi-automatically to stay fresh.
// It is exposed by a get request to /rec/doc/
func generateDoc(w http.ResponseWriter, r *http.Request) {
	title := "rec doc"

	processIn := rec.ProcessInput{
		UserName: "string",
		Audio: rec.Audio{
			FileType: "string",
			Data:     "string of base64 encoded data",
		},
		Text:        "string",
		RecordingID: "string",
	}

	processInSample := rec.ProcessInput{
		ScriptName: "example_proj",
		UserName:   "user0001",
		Audio: rec.Audio{
			FileType: "audio/webm",
			Data:     "GkXfo59ChoEBQ ...",
		},
		Text:        "text to be spoken",
		RecordingID: "utterance_0001",
	}

	prettyJSON, err := rec.PrettyMarshal(processIn)
	if err != nil {
		msg := fmt.Sprintf("failed to pretty marshal : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	prettySampleJSON, err := rec.PrettyMarshal(processInSample)
	if err != nil {
		msg := fmt.Sprintf("failed to pretty marshal : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	processResp := rec.ProcessResponse{
		Ok:                true,
		Confidence:        0.0,
		RecognitionResult: "string",
		RecordingID:       "string",
		Message:           "string",
	}
	prettyResponseJSON, err := rec.PrettyMarshal(processResp)
	if err != nil {
		msg := fmt.Sprintf("failed to pretty marshal : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	t, err := template.New("webpage").Parse(docTemplate)
	if err != nil {
		msg := fmt.Sprintf("failed to parse doc template : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	s00 := `Available server request URLs
               (auto generated from the router):`
	s0 := strings.Join(walkedURLs, "\n")

	s1 := string(prettyJSON)
	s2 := string(prettySampleJSON)
	s3 := string(prettyResponseJSON)
	//s3 := string(prettyResp)
	//_ = s3

	//log.Println(s1)
	//log.Println(s2)
	d := tplData{
		Title: title,
		Items: []item{
			{Desc: s00, Example: s0},
			{Desc: "", Example: "__________________________________________________"},
			{Desc: "Input JSON to POST request to /rec/process/:", Example: s1},
			{Desc: "Sample JSON:", Example: s2},
			{Desc: "", Example: "__________________________________________________"},
			{Desc: "The JSON returned by a successful POST request to /rec/process/: ", Example: s3},
		},
	}

	t.Execute(w, d)
}
