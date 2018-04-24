package adapters

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	u "net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

func RunTensorflowFromURL(rc config.Recogniser, wavFilePath string, input rec.ProcessInput) (rec.RecogniserResponse, error) {
	name := rc.LongName()
	res := rec.RecogniserResponse{RecordingID: input.RecordingID, Source: rc.LongName()}

	wavFilePathAbs, err := filepath.Abs(wavFilePath)
	if err != nil {
		msg := fmt.Sprintf("failed to get absolut path for wav file : %v\n", err)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Status = false
		res.Message = msg
		return res, fmt.Errorf("[%s] %s", name, msg)
	}
	wavFilePathAbs = u.PathEscape(wavFilePathAbs)
	url := strings.Replace(rc.Cmd, wavFilePlaceHolder, wavFilePathAbs, -1)
	log.Printf("runTensorflowFromURL url=%s\n", url)
	log.Printf("runTensorflowFromURL wav rel=%s\n", wavFilePathAbs)

	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("failed to call URL : %v", err)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}
	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("failed to call URL %s : %s", url, resp.Status)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to call URL : %v", err)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	result := strings.TrimSpace(string(body))
	var text string
	res0 := strings.Split(result, "\t")
	if len(res0) < 2 {
		fmt.Println("??? " + result)
		text = "FAIL"
	} else {
		text = res0[1]
	}
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
	log.Printf("runTensorflowFromURL RecognitionResult: %s\n", res.RecognitionResult)
	return res, nil
}
