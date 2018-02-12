package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func runExternalKaldiDecoder(wavFilePath string, res processResponse) (processResponse, error) {

	_, pErr := exec.LookPath("python")
	if pErr != nil {
		log.Printf("failure : %v\n", pErr)
		return res, fmt.Errorf("failed to find the external 'python' command : %v", pErr)
	}

	cmd := exec.Command("python", "decode_test.py", wavFilePath)
	var out bytes.Buffer
	//var sterr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr //&sterr

	err := cmd.Run()
	if err != nil {
		log.Printf("failure: %v\n", err /*sterr.String()*/)
		return res, fmt.Errorf("runExternalKaldiDecoder failed running '%s': %v\n", cmd.Path, err)

	}

	log.Printf("RecognitionResult: %s\n", out.String())
	res.RecognitionResult = strings.TrimSpace(out.String())
	res.Message = "Recognised by external kaldi recognizer"
	return res, nil
}
