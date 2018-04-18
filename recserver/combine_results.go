package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

func isChar(s string) bool {
	return len([]rune(s)) == 1
}

func isWord(s string) bool {
	return !isChar(s)
}

func getUserWeight(input rec.ProcessInput, res rec.ProcessResponse) float32 {
	rcName := res.Source()
	if w, ok := input.Weights[rcName]; ok {
		return w
	}
	return 1.0
}

func getConfigWeight(input rec.ProcessInput, res rec.ProcessResponse, recName2Weights map[string]config.Recogniser) (float32, error) {
	rc, ok := recName2Weights[res.Source()]
	if !ok {
		return -1.0, fmt.Errorf("no recogniser configured for %s", res.Source())
	}
	if w, ok := rc.Weights[input.Text]; ok {
		return w, nil
	} else if w, ok := rc.Weights[res.RecognitionResult]; ok {
		return w, nil
	} else if w, ok := rc.Weights["char"]; ok && isChar(input.Text) {
		return w, nil
	} else if w, ok := rc.Weights["word"]; ok && isWord(input.Text) {
		return w, nil
	} else if w, ok := rc.Weights["default"]; ok {
		return w, nil
	}
	return 1.0, nil
}

// TODO: UNIT TESTS FOR THIS ALGORITHM!
func combineResults(input rec.ProcessInput, inputResults []rec.ProcessResponse, includeOriginalResponses bool) (rec.ProcessResponse, error) {
	var resErr error
	var results = inputResults
	var recName2Weights = make(map[string]config.Recogniser)
	var res2Freq = make(map[string]int)
	for _, rc := range config.MyConfig.Recognisers {
		if !rc.Disabled {
			recName2Weights[rc.LongName()] = rc
		}
	}
	for _, res := range results {
		res2Freq[res.RecognitionResult] += 1
	}
	//var wConf = make(map[string]float32)
	for i, res := range results {
		conf := res.Confidence
		if conf < 0.0 { // confidence below zero => confidence unknown/undefined; confidence zero => kept as is
			conf = 1.0
		}
		configWeight, err := getConfigWeight(input, res, recName2Weights)
		if err != nil {
			resErr = err
		}
		userWeight := getUserWeight(input, res)
		freq := res2Freq[res.RecognitionResult]
		freqNormed := float32(freq) / float32(len(results))
		wc := conf * freqNormed * configWeight * userWeight
		//wConf[res.Source()] = wc
		res.Confidence = wc
		results[i] = res // hmm
		log.Printf("combineResults [%s] '%s' | c=%f f=%f cw=%f uw=%f => %f", res.Source(), res.RecognitionResult, conf, freqNormed, configWeight, userWeight, wc)
	}

	sorter := func(i, j int) bool {
		if results[i].Ok != results[j].Ok {
			return results[i].Ok
		}
		return results[i].Confidence > results[j].Confidence
	}
	sort.Slice(results, sorter)
	if resErr != nil {
		return rec.ProcessResponse{}, nil
	}

	var selected rec.ProcessResponse
	if len(results) > 0 {
		selected = results[0]
		selected.Message = ""
	} else {
		selected = rec.ProcessResponse{Ok: false,
			RecordingID:       input.RecordingID,
			Message:           "No result from server",
			RecognitionResult: ""}
	}
	if includeOriginalResponses {
		selected.ComponentResults = results
	}
	return selected, nil
}
