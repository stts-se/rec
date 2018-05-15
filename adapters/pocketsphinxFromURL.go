package adapters

import (
	// "bytes"
	"fmt"
	"log"
	//"os"
	"encoding/json"
	"io/ioutil"
	"net/http"
	u "net/url"
	"path/filepath"
	"strings"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

type sphinxResp struct {
	RecognisedUtterance string `json:"recognised_utterance"`
}

func pocketsphinxMapElexia(s0 string) string {
	s := strings.TrimSpace(strings.Replace(s0, ".", "", -1))
	if s == "" {
		return "_silence_"
	}
	nWds := len(strings.Split(s, " "))
	if nWds > 2 {
		return "_other_"
	}
	return s
}

func RunElexiaPocketsphinxFromURL(rc config.Recogniser, wavFilePath string, input rec.ProcessInput) (rec.RecogniserResponse, error) {
	res, err := RunPocketsphinxFromURL(rc, wavFilePath, input)
	if err != nil {
		return res, err
	}
	recRes := res.RecognitionResult
	text := pocketsphinxMapElexia(recRes)
	if recRes != "" && text != recRes {
		res.RecognitionResult = text
		res.Message = fmt.Sprintf("original result: %s", recRes)
	}
	return res, err
}

func RunPocketsphinxFromURL(rc config.Recogniser, wavFilePath string, input rec.ProcessInput) (rec.RecogniserResponse, error) {
	name := rc.LongName()
	url := rc.Cmd
	res := rec.RecogniserResponse{RecordingID: input.RecordingID, Source: rc.LongName()}

	if !strings.Contains(url, wavFilePlaceHolder) {
		msg := fmt.Sprintf("input command must contain wav file variable %s", wavFilePlaceHolder)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Status = false
		res.Message = msg
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	wavFilePathAbs, err := filepath.Abs(wavFilePath)
	if err != nil {
		msg := fmt.Sprintf("failed to get absolut path for wav file : %v\n", err)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Status = false
		res.Message = msg
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	//sphinxURL := "http://localhost:8000/rec?audio_file=" + wavFielPathAbs

	// TODO Would it be better to do u.Parse(URL); url = u.EscapedPath() ?
	wavFilePathAbs = u.PathEscape(wavFilePathAbs)
	sphinxURL := strings.Replace(url, wavFilePlaceHolder, wavFilePathAbs, -1)

	log.Printf("callExternalPocketsphinxDecoderServer url=%s\n", sphinxURL)
	log.Printf("callExternalPocketsphinxDecoderServer wav=%s\n", wavFilePathAbs)
	resp, err := http.Get(sphinxURL)
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

	sr := sphinxResp{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read response : %v", err)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	err = json.Unmarshal(body, &sr)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal JSON : %v", err)
		log.Printf("[%s] failure : %s\n", name, msg)
		res.Message = msg
		res.Status = false
		return res, fmt.Errorf("[%s] %s", name, msg)
	}

	recRes := strings.TrimSpace(sr.RecognisedUtterance)

	res.Status = true
	res.RecognitionResult = recRes
	res.Confidence = 1.0
	log.Printf("[%s] RecognitionResult: %s\n", name, res.RecognitionResult)
	return res, nil
}
