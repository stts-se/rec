package main

import (
	"bytes"
	"fmt"
	"log"
	//"os"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stts-se/rec"
)

func runExternalPocketsphinxDecoder(wavFilePath string, input rec.ProcessInput) (rec.ProcessResponse, error) {

	methodName := "pocketsphinx"
	res := rec.ProcessResponse{RecordingID: input.RecordingID}

	_, pErr := exec.LookPath("python")
	if pErr != nil {
		log.Printf("failure : %v\n", pErr)
		return res, fmt.Errorf("failed to find the external 'python' command : %v", pErr)
	}

	cmd := exec.Command("python3", "/home/hanna/git_repos/e-lexia/pocketsphinx/demo_client.py", wavFilePath)
	var out bytes.Buffer
	var sterr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &sterr

	err := cmd.Run()
	if err != nil {
		log.Printf("failure: %v\n", err /*sterr.String()*/)
		log.Printf("stderr: %v", sterr.String())
		return res, fmt.Errorf("runExternalPocketsphinxDecoder failed running '%s': %v\n", cmd.Path, err)

	}

	log.Printf("RecognitionResult: %s\n", out.String())
	text := strings.TrimSpace(out.String())
	if len(text) > 0 {
		res.RecognitionResult = text
		res.Ok = true
	} else {
		res.Ok = false
	}
	msg := "Recognised by external pocketsphinx recognizer"
	res.Message = fmt.Sprintf("[%s] %s", methodName, msg)
	return res, nil
}

type sphinxResp struct {
	recognisedUtterance string `json:"recognised_utterance"`
}

func callExternalPocketsphinxDecoderServer(wavFilePath string, input rec.ProcessInput) (rec.ProcessResponse, error) {

	methodName := "pocketsphinx"
	res := rec.ProcessResponse{RecordingID: input.RecordingID}

	cd, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		//log.Fatal(err)
		return res, fmt.Errorf("failed to get path to current dir : %v", err)
	}

	sphinxURL := "http://localhost:8000/rec?audio_file=" + filepath.Join(cd, wavFilePath)
	resp, err := http.Get(sphinxURL)
	if err != nil {
		return res, fmt.Errorf("callExternalPocketsphinxDecoderServer: failed get '%s' : %v", sphinxURL, err)
	}

	sr := sphinxResp{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, fmt.Errorf("callExternalPocketsphinxDecoderServer: failed to read response : %v", err)
	}

	err = json.Unmarshal(body, &sr)
	if err != nil {
		return res, fmt.Errorf("callExternalPocketsphinxDecoderServer: failed to unmarshal JSON '%s' : %v", string(body), err)
	}

	recRes := sr.recognisedUtterance

	log.Printf("RecognitionResult: %s\n", recRes)
	text := strings.TrimSpace(recRes)
	if len(text) > 0 {
		res.RecognitionResult = text
		res.Ok = true
	} else {
		res.Ok = false
	}
	msg := "Recognised by external pocketsphinx recognizer"
	res.Message = fmt.Sprintf("[%s] %s", methodName, msg)
	return res, nil
}
