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
func combineResults(input rec.ProcessInput, inputResults []rec.ProcessResponse, includeOriginalResponses bool) (rec.ProcessResponse, error) {
	results := inputResults
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
	for i, res := range results {
		rc, ok := recName2Weights[res.Source()]
		if !ok {
			return rec.ProcessResponse{}, fmt.Errorf("no recogniser configured for %s", res.Source())
		}
		var weight float32 = 1.0
		if w, ok := rc.Weights[input.Text]; ok {
			weight = w

		} else if w, ok := rc.Weights[res.RecognitionResult]; ok {
			weight = w

		} else if w, ok := rc.Weights["char"]; ok && isChar(input.Text) {
			weight = w

		} else if w, ok := rc.Weights["word"]; ok && isWord(input.Text) {
			weight = w
		} else {
			weight = rc.Weights["default"]
		}
		conf := res.Confidence
		if conf < 0.0 { // confidence below zero => confidence unknown/undefined; confidence zero => kept as is
			conf = 1.0
		}
		freq := res2Freq[res.RecognitionResult]
		freqNormed := float32(freq) / float32(len(results))
		wc := weight * conf * freqNormed
		wConf[res.Source()] = wc
		res.Confidence = wc
		results[i] = res // hmm
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

	var selected rec.ProcessResponse
	if len(results) > 0 {
		selected = results[0]
		wc, ok := wConf[selected.Source()]
		if !ok {
			return selected, fmt.Errorf("no weighted confidence for %s", selected.Source())
		}
		selected.Confidence = wc
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
