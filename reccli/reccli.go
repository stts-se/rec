package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/stts-se/rec"
)

type F struct {
	s string
}

func (f F) String() string { return f.s }
func (f *F) Set(s string) error {
	f.s = s
	return nil
}

var flagUserName, flagRecordingID, flagURL, flagText F

func init() {
	flagUserName = F{s: "tmpuser0"}
	flagRecordingID = F{s: "tmprecid0"}
	flagURL = F{s: "http://localhost:9993/rec/process/"}
	flagText = F{s: "DUMMY_TEXT0"}

	flag.Var(&flagURL, "url", "URL for calling rec server.")
	flag.Var(&flagUserName, "u", "user name to be sent to server along with sound file.")
	flag.Var(&flagRecordingID, "r", "recording ID to be sent to server along with sound file.")
	flag.Var(&flagText, "t", "text to be sent to server along with sound file.")
}

func main() {

	flag.Parse()
	fileName := flag.Arg(1)
	if fileName == "" {
		fmt.Fprintf(os.Stderr, "reccli <AUDIO FILE> or --help.\n") //, os.Args[0])
		os.Exit(0)
	}

	bts, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read file '%s' : %v\n", fileName, err)
		os.Exit(1)
	}

	aud := base64.StdEncoding.EncodeToString(bts)
	ext := strings.TrimPrefix(path.Ext(path.Base(fileName)), ".")

	payload := rec.ProcessInput{
		UserName:    flagUserName.String(),
		RecordingID: flagRecordingID.String(),
		Text:        flagText.String(),
		Audio: rec.Audio{
			Data:     aud,
			FileType: "audio/" + ext,
		},
	}

	pl, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal JSON : %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(flagURL.String(), "application/json", bytes.NewBuffer(pl))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to call server : %v\n", err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read server response : %v\n", err)
		os.Exit(1)

	}

	if resp.StatusCode != 200 {

		fmt.Fprintf(os.Stderr, "response Status: %s\n", resp.Status)
		fmt.Fprintf(os.Stderr, "response Headers: %s\n", resp.Header)
		fmt.Fprintf(os.Stderr, "response Body: %s\n", string(body))

		os.Exit(1)
	}

	// Pretty print the returned JSON
	var prettyBody bytes.Buffer
	err = json.Indent(&prettyBody, body, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to process JSON '%s': %v\n", string(body), err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "%s\n", string(prettyBody.Bytes()))
}
