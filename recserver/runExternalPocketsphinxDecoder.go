package main

import (
	"bytes"
	"fmt"
	"log"
	//"os"
	"os/exec"
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
