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

func RunTensorflowCommand(rc config.Recogniser, wavFilePath string, input rec.ProcessInput) (rec.RecogniserResponse, error) {
	name := rc.LongName()
	command := rc.Cmd
	res := rec.RecogniserResponse{RecordingID: input.RecordingID, Source: rc.LongName()}

	if !strings.Contains(command, wavFilePlaceHolder) {
		msg := fmt.Sprintf("input command must contain wav file variable %s", wavFilePlaceHolder)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	wavFilePathAbs, err := filepath.Abs(wavFilePath)
	if err != nil {
		msg := fmt.Sprintf("failed to get absolute path for wav file : %v", err)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
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
		msg := fmt.Sprintf("failed running command %v : %v", cmd, stderr.String())
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	// FILE TAB RES TAB SCORE
	outS := out.String()
	res0 := strings.Split(strings.TrimSpace(outS), "\t")
	text := res0[1]
	if text == "FAIL" {
		res.Status = false
	} else {
		res.Status = true
		score, err := strconv.ParseFloat(res0[2], 64)
		if err != nil {
			msg := fmt.Sprintf("failed parsing score to float64 : %v", err)
			log.Printf("[%s] failure : %s\n", name, msg)
			res.Message = msg
			res.Status = false
			return res, fmt.Errorf("[%s] %s", name, msg)
		}
		res.RecognitionResult = text
		res.Confidence = score
	}
	log.Printf("[%s] RecognitionResult: %s\n", name, res.RecognitionResult)
	return res, nil
}
