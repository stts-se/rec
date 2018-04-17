package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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

func runGStreamerKaldiFromURL(rc config.Recogniser, wavFilePath string, input rec.ProcessInput) (rec.ProcessResponse, error) {
	url := rc.Cmd
	res := rec.ProcessResponse{RecordingID: input.RecordingID}

	// curl -T $WAVFILE "http://192.168.0.105:8080/client/dynamic/recognise"
	// {"status": 0, "hypotheses": [{"utterance": "just three style."}], "id": "80a4a3e6-15ec-41e7-ac5d-fa2ea2386df2"}

	log.Printf("runGStreamerKaldiFromURL url=%s\n", url)
	log.Printf("runGStreamerKaldiFromURL wav rel=%s\n", wavFilePath)

	audio, err := ioutil.ReadFile(wavFilePath)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("failed to read audio into byte array : %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(audio))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("failed to send post request : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("failed to run kaldi gstreamer, got %s", resp.Status)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	gsResp := gstreamerResponse{}
	err = json.Unmarshal(body, &gsResp)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("failed to unmarshal : %v", err)
	}

	if len(gsResp.Hypotheses) > 0 {
		res.RecognitionResult = gsResp.Hypotheses[0].Utterance
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
