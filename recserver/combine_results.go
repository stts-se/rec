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

// TODO: UNIT TESTS FOR THIS ALGORITHM!
func combineResults(input rec.ProcessInput, results []rec.ProcessResponse) (rec.ProcessResponse, error) {
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

	//log.Printf("res2Freq: %#v\n", res2Freq)

	var err error

	applyWeights := func(res rec.ProcessResponse) float32 {
		rc, ok := recName2Weights[res.Source()]
		if !ok {
			err = fmt.Errorf("no recogniser configured for %s", res.Source())
		}
		conf := res.Confidence
		if conf == 0.0 {
			conf = 0.65 // default
		}
		weight := rc.Weights["default"]
		if w, ok := rc.Weights["char"]; ok && isChar(input.Text) {
			weight = w

		} else if w, ok := rc.Weights["word"]; ok && isWord(input.Text) {
			weight = w
		}
		freq := res2Freq[res.RecognitionResult]
		wConf := weight * conf * float32(freq)
		log.Printf("combineResults [%s] '%s' - w=%f c=%f f=%d => %f", res.Source(), res.RecognitionResult, weight, conf, freq, wConf)
		return wConf
	}

	if err != nil {
		return rec.ProcessResponse{}, nil
	}

	sorter := func(i, j int) bool {
		wI := applyWeights(results[i])
		wJ := applyWeights(results[j])

		if results[i].Ok != results[j].Ok {
			return results[i].Ok
		} else {
			return wI > wJ
		}
	}
	sort.Slice(results, sorter)
	var r1 rec.ProcessResponse
	if len(results) > 0 {
		r1 = results[0]
		r1.Message = ""
	} else {
		r1 = rec.ProcessResponse{Ok: false,
			RecordingID:       input.RecordingID,
			Message:           "No result from server",
			RecognitionResult: ""}
	}
	return r1, nil
}
