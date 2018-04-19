package main

import (
	"fmt"
	"log"
	"math"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

func isChar(s string) bool {
	return len([]rune(s)) == 1
}

func isWord(s string) bool {
	return !isChar(s)
}

func roundConfidence(fl float64) float64 {
	unit := 0.0001
	return math.Round(fl/unit) * unit
}

func getUserWeight(input rec.ProcessInput, res rec.ProcessResponse) float64 {
	rcName := res.Source()
	if w, ok := input.Weights[rcName]; ok {
		return w
	}
	return 1.0
}

func getConfigWeight(input rec.ProcessInput, res rec.ProcessResponse, recName2Weights map[string]config.Recogniser) (float64, error) {
	rc, ok := recName2Weights[res.Source()]
	if !ok {
		msg := fmt.Sprintf("no recogniser defined for %s", res.Source())
		return 0.0, fmt.Errorf("%s", msg)
	}
	ws := rc.Weights
	if w, ok := ws[input.Text]; ok {
		return w, nil
		// } else if w, ok := ws[res.RecognitionResult]; ok {
		// 	return w
	} else if w, ok := ws["char"]; ok && (isChar(input.Text)) { // || isChar(res.RecognitionResult)) {
		return w, nil
	} else if w, ok := ws["word"]; ok && (isWord(input.Text)) { // || isWord(res.RecognitionResult)) {
		return w, nil
	} else if w, ok := ws["default"]; ok {
		return w, nil
	}
	return 1.0, nil
}

func getBestGuess(totalConfs map[string]float64) (string, float64) {
	var bestConf = -1.0
	var bestGuess string
	for guess, conf := range totalConfs {
		if conf > bestConf {
			bestConf = conf
			bestGuess = guess
		}
	}
	return bestGuess, bestConf
}

// TODO: UNIT TESTS FOR THIS ALGORITHM!
func combineResults(input rec.ProcessInput, inputResults []rec.ProcessResponse, includeOriginalResponses bool) (rec.ProcessResponse, error) {
	var resErr error
	var results = inputResults
	var recName2Weights = make(map[string]config.Recogniser)
	for _, rc := range config.MyConfig.EnabledRecognisers() {
		if !rc.Disabled {
			recName2Weights[rc.LongName()] = rc
		}
	}

	// compute initial weights (recogniser conf * config defined conf * user defined conf)
	var totalConf = 0.0
	for i, res := range results {
		inputConf := res.Confidence // input confidence from recogniser
		if inputConf < 0.0 {        // below zero => unknown/undefined => default value 1.0
			inputConf = 1.0
		}
		configWeight, err := getConfigWeight(input, res, recName2Weights)
		if err != nil {
			return rec.ProcessResponse{}, fmt.Errorf("combineResults failed : %v", err)
		}
		userWeight := getUserWeight(input, res)
		intermWeight := inputConf * configWeight * userWeight // intermediate weight
		if !res.Ok {
			intermWeight = 0.0
		}
		res.Confidence = intermWeight
		totalConf += roundConfidence(res.Confidence)
		results[i] = res // update the slice with new value
		log.Printf("combineResults:1 [%s] '%s' | conf=%f rw=%f uw=%f => %f", res.Source(), res.RecognitionResult, inputConf, configWeight, userWeight, intermWeight)
	}
	// re-compute conf relative to the sum of weights
	var totalConfs = make(map[string]float64) // result string => sum of confidence measures for responses with this result
	for i, res := range results {
		newConf := roundConfidence(res.Confidence / totalConf)
		res.Confidence = newConf
		results[i] = res // update the slice with new value
		totalConfs[res.RecognitionResult] = totalConfs[res.RecognitionResult] + res.Confidence
		log.Printf("combineResults:2 [%s] '%s' | conf=%f", res.Source(), res.RecognitionResult, res.Confidence)
	}

	if resErr != nil {
		return rec.ProcessResponse{}, nil
	}

	var selected rec.ProcessResponse
	if len(results) > 0 {
		bestGuess, weight := getBestGuess(totalConfs)
		selected = rec.ProcessResponse{Ok: true,
			RecordingID:       input.RecordingID,
			Message:           "",
			RecognitionResult: bestGuess,
			Confidence:        weight}
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
