package adapters

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

type tensorflowResponse struct {
	Status     int    `json:"status"`
	Hypotheses []hypo `json:"hypotheses"`
	Id         string `json:"id"`
	Message    string `json:"message"`
}

var wavFilePlaceHolder = "{wavfile}"

func RunTensorflowCommand(rc config.Recogniser, wavFilePath string, input rec.ProcessInput) (rec.ProcessResponse, error) {
	name := rc.LongName()
	command := rc.Cmd
	res := rec.ProcessResponse{RecordingID: input.RecordingID}

	if !strings.Contains(command, wavFilePlaceHolder) {
		msg := fmt.Sprintf("[%s] input command must contain wav file variable %s", name, wavFilePlaceHolder)
		log.Printf("failure : %v\n", msg)
		return res, fmt.Errorf(msg)
	}

	wavFilePathAbs, err := filepath.Abs(wavFilePath)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("[%s] failed to get absolut path for wav file : %v", name, err)
	}

	command = strings.Replace(command, wavFilePlaceHolder, wavFilePathAbs, -1)

	cmdSplit := strings.Fields(command)
	cmd := exec.Command(cmdSplit[0], cmdSplit[1:]...)

	command = strings.Replace(command, wavFilePlaceHolder, wavFilePathAbs, -1)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Printf(stderr.String())
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("[%s] failed running command %v : %v", name, cmd, err)
	}

	// FILE TAB RES TAB SCORE
	outS := out.String()
	res0 := strings.Split(strings.TrimSpace(outS), "\t")
	text := res0[1]
	if text == "FAIL" {
		res.Ok = false
	} else {
		score, err := strconv.ParseFloat(res0[2], 64)
		if err != nil {
			log.Printf("failure : %v\n", err)
			return res, fmt.Errorf("[%s] failed parsing score to float64 : %v", name, err)
		}
		res.Ok = true
		res.RecognitionResult = text
		res.Confidence = score
	}

	res.Message = rc.LongName()
	log.Printf("[%s] RecognitionResult: %s\n", name, res.RecognitionResult)
	return res, nil
}
