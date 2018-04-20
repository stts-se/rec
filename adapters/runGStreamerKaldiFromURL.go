package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

type hypo struct {
	Utterance string `json:"utterance"`
}

type gstreamerResponse struct {
	Status     int    `json:"status"`
	Hypotheses []hypo `json:"hypotheses"`
	Id         string `json:"id"`
	Message    string `json:"message"`
}

var gStreamerENMaptable = map[string]string{
	"a":        "a",
	"ace":      "is",
	"all said": "rose",
	"all":      "o",
	"also":     "rose",
	"and also": "rose",
	"b":        "bi",
	"be":       "bi",
	"e":        "i",
	"each":     "is",
	"he's":     "is",
	"is":       "is",
	"most":     "mos",
	"o":        "o",
	"place":    "blæs",
	"small":    "sne",
	"the":      "bi",
	"yes":      "blæs",
}

func gStreamerENMapText(s0 string) (string, float64) {
	s := strings.TrimSpace(strings.Replace(s0, ".", "", -1))
	if s == "" {
		return "_silence_", 1.0
	}
	nWds := len(strings.Split(s, " "))
	if nWds > 2 {
		return "_other_", 2.0
	}
	if mapped, ok := gStreamerENMaptable[s]; ok {
		return mapped, 1.0
	}
	return s, 0.0
}

func RunGStreamerKaldiFromURL(rc config.Recogniser, wavFilePath string, input rec.ProcessInput) (rec.ProcessResponse, error) {
	name := rc.LongName()
	url := rc.Cmd
	res := rec.ProcessResponse{RecordingID: input.RecordingID}

	// curl -T $WAVFILE "http://192.168.0.105:8080/client/dynamic/recognize"
	// {"status": 0, "hypotheses": [{"utterance": "just three style."}], "id": "80a4a3e6-15ec-41e7-ac5d-fa2ea2386df2"}

	log.Printf("runGStreamerKaldiFromURL url=%s\n", url)
	log.Printf("runGStreamerKaldiFromURL wav rel=%s\n", wavFilePath)

	audio, err := ioutil.ReadFile(wavFilePath)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("[%s] failed to read audio into byte array : %v", name, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(audio))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("[%s] failed to send post request : %v", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("[%s] failed to run kaldi gstreamer, got %s", name, resp.Status)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	gsResp := gstreamerResponse{}
	err = json.Unmarshal(body, &gsResp)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("[%s] failed to unmarshal : %v", name, err)
	}

	if len(gsResp.Hypotheses) > 0 {
		res0 := strings.ToLower(gsResp.Hypotheses[0].Utterance)
		newRes, conf := gStreamerENMapText(res0)
		res.Confidence = conf
		res.RecognitionResult = newRes
		res.Ok = true
	} else {
		res.RecognitionResult = ""
		res.Ok = false
	}
	if gsResp.Status != 0 {
		res.Ok = false
	}
	res.Message = rc.LongName()
	log.Printf("runGStreamerKaldiFromURL RecognitionResult: %s\n", res.RecognitionResult)
	return res, nil
}
