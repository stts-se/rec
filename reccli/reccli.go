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
	"strconv"
	"strings"

	"github.com/stts-se/rec"
)

func plPretty(pl0 rec.ProcessInput) string {
	pl := pl0
	pl.Audio.Data = ""
	bytes, err := json.MarshalIndent(pl, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to process rec.ProcessInput '%v': %v\n", pl, err)
		os.Exit(1)
	}
	return string(bytes)
}

func main() {
	var cmdName = "reccli"
	var flagUserName, flagRecordingID, flagURL, flagText, flagWeights string
	flag.StringVar(&flagURL, "url", "http://localhost:9993/rec/process/?verb=true", "URL for calling rec server.")
	flag.StringVar(&flagUserName, "u", "tmpuser0", "user name to be sent to server along with sound file.")
	flag.StringVar(&flagRecordingID, "r", "tmprecid0", "recording ID to be sent to server along with sound file.")
	flag.StringVar(&flagText, "t", "DUMMY_TEXT0", "text to be sent to server along with sound file.")
	flag.StringVar(&flagWeights, "w", "", "user defined weights for recognisers (& separated list: NAME1=WEIGHT1&NAME2=WRIGHT2).")

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

		fmt.Fprintf(os.Stdout, "[%s] AUDIO %s\n", cmdName, fileName)

		aud := base64.StdEncoding.EncodeToString(bts)
		ext := strings.TrimPrefix(path.Ext(path.Base(fileName)), ".")

		weights := make(map[string]float64)
		if len(flagWeights) > 0 {
			for _, w0 := range strings.Split(flagWeights, "&") {
				x := strings.Split(w0, "=")
				if len(x) != 2 {
					fmt.Fprintf(os.Stderr, "couldn't parse input weights %s\n", flagWeights)
					os.Exit(1)
				}
				rcName := x[0]
				w, err := strconv.ParseFloat(x[1], 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "couldn't parse input weights %s : %v\n", flagWeights, err)
					os.Exit(1)
				}
				weights[rcName] = w
			}
		}

		payload := rec.ProcessInput{
			UserName:    flagUserName,
			RecordingID: flagRecordingID,
			Text:        flagText,
			Audio: rec.Audio{
				Data:     aud,
				FileType: "audio/" + ext,
			},
			Weights: weights,
		}

		fmt.Fprintf(os.Stdout, "[%s] INPUT %s\n", cmdName, plPretty(payload))

		pl, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal JSON : %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "[%s] URL %s\n", cmdName, flagURL)
		resp, err := http.Post(flagURL, "application/json", bytes.NewBuffer(pl))
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

		fmt.Fprintf(os.Stdout, "[%s] RESPONSE %s\n", cmdName, string(prettyBody.Bytes()))
	}
}
