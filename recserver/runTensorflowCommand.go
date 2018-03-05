package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type tensorflowResponse struct {
	Status     int    `json:"status"`
	Hypotheses []hypo `json:"hypotheses"`
	Id         string `json:"id"`
	Message    string `json:"message"`
}

var wavFilePlaceHolder = "{wavfile}"

var scoreRe = regexp.MustCompile("^([^ ]+) [(]score = ([0-9.]+)[)]$")

// to activate tensorflow:
// source ~/progz/tensorflow/bin/activate

func runTensorflowCommand(command string, wavFilePath string, res processResponse) (processResponse, error) {

	methodName := "tensorflow"

	if !strings.Contains(command, wavFilePlaceHolder) {
		msg := fmt.Sprintf("input tensorflow command must contain wav file variable %s", wavFilePlaceHolder)
		log.Printf("failure : %v\n", msg)
		return res, fmt.Errorf(msg)
	}

	wavFilePathAbs, err := filepath.Abs(wavFilePath)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("failed to get absolut path for wav file : %v", err)
	}

	command = strings.Replace(command, wavFilePlaceHolder, wavFilePathAbs, -1)

	log.Printf("runTensorflowCommand cmd=%s\n", command)

	cmdSplit := strings.Fields(command)
	cmd := exec.Command(cmdSplit[0], cmdSplit[1:]...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Printf(stderr.String())
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("failed running command %v : %v", cmd, err)
	}

	// <WORD> (score = <SCORE>)
	res0 := strings.Split(out.String(), "\n")[0]
	match := scoreRe.FindStringSubmatch(res0)
	if len(match) != 3 {
		msg := fmt.Sprintf("couldn't parse match result '%s' into text and score", res0)
		log.Printf("failure : %s\n", msg)
		return res, fmt.Errorf("%s", msg)
	}
	text := match[1]
	score, err := strconv.ParseFloat(match[2], 64)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("failed parsing score to float64 : %v", err)
	}
	res.RecognitionResult = text
	res.Confidence = float32(score)

	msg := "Recognised by external tensorflow recognizer"

	if len(res.Message) > 0 {
		res.Message = res.Message + "; " + fmt.Sprintf("[%s] %s", methodName, msg)
	} else {
		res.Message = fmt.Sprintf("[%s] %s", methodName, msg)
	}

	log.Printf("runTensorflowCommand RecognitionResult: %s\n", res.RecognitionResult)
	return res, nil
}
