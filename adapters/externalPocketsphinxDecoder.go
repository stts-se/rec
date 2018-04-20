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

// func runExternalPocketsphinxDecoder(wavFilePath string, input rec.ProcessInput) (rec.ProcessResponse, error) {

// 	panic("runExternalPocketsphinxDecoder is deprecated")

// 	methodName := "pocketsphinx"
// 	res := rec.ProcessResponse{RecordingID: input.RecordingID}

// 	_, pErr := exec.LookPath("python")
// 	if pErr != nil {
// 		log.Printf("[%s] failure : %v\n", name, pErr)
// 		return res, fmt.Errorf("failed to find the external 'python' command : %v", pErr)
// 	}

// 	cmd := exec.Command("python3", "/home/hanna/git_repos/e-lexia/pocketsphinx/demo_client.py", wavFilePath)
// 	var out bytes.Buffer
// 	var sterr bytes.Buffer
// 	cmd.Stdout = &out
// 	cmd.Stderr = &sterr

// 	err := cmd.Run()
// 	if err != nil {
// 		log.Printf("failure: %v\n", err /*sterr.String()*/)
// 		log.Printf("stderr: %v", sterr.String())
// 		return res, fmt.Errorf("runExternalPocketsphinxDecoder failed running '%s': %v\n", cmd.Path, err)

// 	}

// 	log.Printf("RecognitionResult: %s\n", out.String())
// 	text := strings.TrimSpace(out.String())
// 	if len(text) > 0 {
// 		res.RecognitionResult = text
// 		res.Ok = true
// 	} else {
// 		res.Ok = false
// 	}
// 	msg := "Recognised by external pocketsphinx recogniser"
// 	res.Message = fmt.Sprintf("[%s] %s", methodName, msg)
// 	return res, nil
// }

type sphinxResp struct {
	RecognisedUtterance string `json:"recognised_utterance"`
}

func CallExternalPocketsphinxDecoderServer(rc config.Recogniser, wavFilePath string, input rec.ProcessInput) (rec.SubProcessResponse, error) {
	name := rc.LongName()
	url := rc.Cmd
	res := rec.SubProcessResponse{RecordingID: input.RecordingID, Source: rc.LongName()}

	if !strings.Contains(url, wavFilePlaceHolder) {
		msg := fmt.Sprintf("[%s] input command must contain wav file variable %s", name, wavFilePlaceHolder)
		log.Printf("[%s] failure : %v\n", name, msg)
		res.Message = "SERVER ERROR"
		return res, fmt.Errorf(msg)
	}

	wavFilePathAbs, err := filepath.Abs(wavFilePath)
	if err != nil {
		log.Printf("[%s] failure : %v\n", name, err)
		res.Message = "SERVER ERROR"
		return res, fmt.Errorf("[%s] failed to get absolut path for wav file : %v", name, err)
	}

	//sphinxURL := "http://localhost:8000/rec?audio_file=" + wavFielPathAbs

	// TODO Would it be better to do u.Parse(URL); url = u.EscapedPath() ?
	wavFilePathAbs = u.PathEscape(wavFilePathAbs)
	sphinxURL := strings.Replace(url, wavFilePlaceHolder, wavFilePathAbs, -1)

	log.Printf("callExternalPocketsphinxDecoderServer URL: %s\n", sphinxURL)
	resp, err := http.Get(sphinxURL)
	if err != nil {
		res.Message = "SERVER ERROR"
		return res, fmt.Errorf("[%s] failed get '%s' : %v", name, sphinxURL, err)
	}

	sr := sphinxResp{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		res.Message = "SERVER ERROR"
		return res, fmt.Errorf("[%s] failed to read response : %v", name, err)
	}

	err = json.Unmarshal(body, &sr)
	if err != nil {
		res.Message = "SERVER ERROR"
		return res, fmt.Errorf("[%s] failed to unmarshal JSON '%s' : %v", name, string(body), err)
	}

	recRes := sr.RecognisedUtterance

	text := strings.TrimSpace(recRes)
	if len(text) == 0 {
		text = "_silence_"
	}
	if text == input.Text {
		res.Ok = true
	} else {
		res.Ok = false
	}

	res.RecognitionResult = text
	res.Confidence = -1.0
	log.Printf("[%s] RecognitionResult: %s\n", name, res.RecognitionResult)
	return res, nil
}
