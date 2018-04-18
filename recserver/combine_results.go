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
	var wConf = make(map[string]float32)
	for _, res := range results {
		rc, ok := recName2Weights[res.Source()]
		if !ok {
			return rec.ProcessResponse{}, fmt.Errorf("no recogniser configured for %s", res.Source())
		}
		var weight = rc.Weights["default"]
		if w, ok := rc.Weights[input.Text]; ok {
			weight = w

		} else if w, ok := rc.Weights[res.RecognitionResult]; ok {
			weight = w

		} else if w, ok := rc.Weights["char"]; ok && isChar(input.Text) {
			weight = w

		} else if w, ok := rc.Weights["word"]; ok && isWord(input.Text) {
			weight = w
		}
		conf := res.Confidence
		if conf < 0.0 { // confidence below zero => confidence unknown/undefined; confidence zero => kept as is
			conf = 0.65
		}
		freq := res2Freq[res.RecognitionResult]
		freqNormed := float32(freq) / float32(len(results))
		wc := weight * conf * freqNormed
		wConf[res.Source()] = wc
		log.Printf("combineResults [%s] '%s' | w=%f c=%f f=%f => %f", res.Source(), res.RecognitionResult, weight, conf, freqNormed, wc)
	}

	var err error

	sorter := func(i, j int) bool {
		if results[i].Ok != results[j].Ok {
			return results[i].Ok
		} else {
			wI, ok := wConf[results[i].Source()]
			if !ok {
				err = fmt.Errorf("no weighted confidence for %s", results[i].Source())
			}
			wJ, ok := wConf[results[j].Source()]
			if !ok {
				err = fmt.Errorf("no weighted confidence for %s", results[j].Source())
			}

			return wI > wJ
		}
	}
	sort.Slice(results, sorter)
	if err != nil {
		return rec.ProcessResponse{}, nil
	}

	var r1 rec.ProcessResponse
	if len(results) > 0 {
		r1 = results[0]
		wc, ok := wConf[r1.Source()]
		if !ok {
			return r1, fmt.Errorf("no weighted confidence for %s", r1.Source())
		}
		r1.Confidence = wc
		r1.Message = ""

	} else {
		r1 = rec.ProcessResponse{Ok: false,
			RecordingID:       input.RecordingID,
			Message:           "No result from server",
			RecognitionResult: ""}
	}
	return r1, nil
}
