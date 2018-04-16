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

func main() {
	var flagUserName, flagRecordingID, flagURL, flagText string
	flag.StringVar(&flagURL, "url", "http://localhost:9993/rec/process/", "URL for calling rec server.")
	flag.StringVar(&flagUserName, "u", "tmpuser0", "user name to be sent to server along with sound file.")
	flag.StringVar(&flagRecordingID, "r", "tmprecid0", "recording ID to be sent to server along with sound file.")
	flag.StringVar(&flagText, "t", "DUMMY_TEXT0", "text to be sent to server along with sound file.")

	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "reccli <AUDIO FILE> or --help\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	for _, fileName := range flag.Args() {
		if fileName == "" {
			fmt.Fprintf(os.Stderr, "reccli <AUDIO FILE> or --help\n") //, os.Args[0])
			flag.PrintDefaults()
			//fmt.Fprintf(os.Stderr, "%s\n", flag.PrintDefaults())
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
			UserName:    flagUserName,
			RecordingID: flagRecordingID,
			Text:        flagText,
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

		resp, err := http.Post(flagURL+"?dev=true", "application/json", bytes.NewBuffer(pl))
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

		fmt.Fprintf(os.Stdout, "%s\n", fileName)
		fmt.Fprintf(os.Stdout, "%s\n", string(prettyBody.Bytes()))
	}
}
