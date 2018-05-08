package adapters

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	u "net/url"
	"path/filepath"
	"strings"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

type tflowResp struct {
	Status              bool    `json:"status"`
	RecognisedUtterance string  `json:"recognised_utterance"`
	Confidence          float64 `json:"confidence"`
	Message             string  `json:"message"`
}

func tensorflowMapText(s0 string) (string, bool) {
	s := strings.TrimSpace(strings.Replace(s0, ".", "", -1))
	if s == "vowel" {
		return "_vowel_", false
	} else if s == "cons" {
		return "_cons_", false
	}
	return s, false
}

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
	log.Printf("runTensorflowFromURL wav=%s\n", wavFilePathAbs)

	resp, err := http.Get(url)
	fmt.Println("RESP=", resp)
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

	tr := tflowResp{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read response : %v", err)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	err = json.Unmarshal(body, &tr)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal JSON : %v", err)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	if tr.Status == false {
		msg := fmt.Sprintf("failed to call URL %s : %s", url, tr.Message)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	recRes := strings.TrimSpace(tr.RecognisedUtterance)

	text, updated := tensorflowMapText(recRes)
	if recRes != "" && text != recRes {
		res.Message = fmt.Sprintf("original result: %s", recRes)
	}
	res.Status = true
	res.RecognitionResult = text
	if recRes != "" && updated && text != recRes {
		msg := fmt.Sprintf("original result: %s", recRes)
		if len(res.Message) > 0 {
			res.Message = res.Message + "; " + msg
		} else {
			res.Message = msg
		}
	}
	res.Message = tr.Message
	res.Confidence = tr.Confidence
	log.Printf("[%s] RecognitionResult: %s\n", name, res.RecognitionResult)
	return res, nil

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	msg := fmt.Sprintf("failed to call URL : %v", err)
	// 	log.Printf("[%s] failure : %s\n", name, msg)
	// 	res.Message = msg
	// 	res.Status = false
	// 	return res, fmt.Errorf("[%s] %s", name, msg)
	// }

	// result := strings.TrimSpace(string(body))
	// // WAVFILE TAB STATUS TAB TEXT TAB CONFIDENCE
	// //   where
	// //    STATUS = FAIL/OK
	// //    TEXT   = RESULT or ERROR MESSAGE
	// log.Printf("runTensorflowFromURL result=%s", result)
	// fields := strings.Split(result, "\t")
	// if len(fields) != 4 {
	// 	msg := fmt.Sprintf("expected four fields back from tensorflow server, found %d : %v", len(fields), fields)
	// 	log.Printf("[%s] failure : %s\n", name, msg)
	// 	res.Message = msg
	// 	res.Status = false
	// 	return res, fmt.Errorf("[%s] %s", name, msg)
	// }
	// status := fields[1]
	// recRes := fields[2]
	// text, updated := tensorflowMapText(recRes)

	// if status == "FAIL" {
	// 	log.Printf("[%s] failure : %s\n", name, text)
	// 	res.Message = text
	// 	res.Status = false
	// 	return res, fmt.Errorf("[%s] %s", name, text)
	// } else if status == "OK" {
	// 	res.Status = true
	// 	score, err := strconv.ParseFloat(fields[3], 64)
	// 	if err != nil {
	// 		msg := fmt.Sprintf("failed parsing score to float64 : %v", err)
	// 		log.Printf("[%s] failure : %s\n", name, msg)
	// 		res.Message = msg
	// 		res.Status = false
	// 		return res, fmt.Errorf("[%s] %s", name, msg)
	// 	}
	// 	res.RecognitionResult = text
	// 	if recRes != "" && updated && text != recRes {
	// 		msg := fmt.Sprintf("original result: %s", recRes)
	// 		if len(res.Message) > 0 {
	// 			res.Message = res.Message + "; " + msg
	// 		} else {
	// 			res.Message = msg
	// 		}
	// 	}
	// 	res.Confidence = score
	// } else {
	// 	msg := fmt.Sprintf("unknown return status %s in %s", status, result)
	// 	log.Printf("[%s] failure : %s\n", name, msg)
	// 	res.Message = msg
	// 	res.Status = false
	// 	return res, fmt.Errorf("[%s] %s", name, msg)
	// }
	// log.Printf("runTensorflowFromURL RecognitionResult: %s\n", res.RecognitionResult)
	// return res, nil
}
