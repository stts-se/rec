package aggregator

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

func isChar(s string) bool {
	return len([]rune(s)) == 1
}

func isWord(s string) bool {
	return len([]rune(s)) > 1
}

func isEntity(s string) bool {
	return strings.HasPrefix(s, "_") && strings.HasSuffix(s, "_")
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
	if w, ok := ws["input:"+input.Text]; ok {
		return w, nil
	} else if w, ok := ws["output:"+res.RecognitionResult]; ok {
		return w, nil
	} else if w, ok := ws["input:_char_"]; ok && isChar(input.Text) {
		return w, nil
	} else if w, ok := ws["input:_word_"]; ok && isWord(input.Text) {
		return w, nil
	} else if w, ok := ws["output:_char_"]; ok && isChar(res.RecognitionResult) {
		return w, nil
	} else if w, ok := ws["output:_word_"]; ok && isWord(res.RecognitionResult) {
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
	if bestConf == 1 { // we can never be 100% sure
		bestConf = 0.99
	}
	return bestGuess, bestConf
}

func CombineResults(input rec.ProcessInput, inputResults []rec.ProcessResponse, includeOriginalResponses bool) (rec.ProcessResponse, error) {
	var resErr error
	var recName2Weights = make(map[string]config.Recogniser)
	for _, rc := range config.MyConfig.EnabledRecognisers() {
		if !rc.Disabled {
			recName2Weights[rc.LongName()] = rc
		}
	}

	// compute initial weights (recogniser conf * config defined conf * user defined conf)
	var totalConf = 0.0
	var anyOk = false
	var convertedResults = []rec.SubProcessResponse{}
	for _, res := range inputResults {
		recogConf := res.Confidence // input confidence from recogniser
		if recogConf < 0.0 {        // below zero => unknown/undefined => default value 1.0
			recogConf = 1.0
		}
		configWeight, err := getConfigWeight(input, res, recName2Weights)
		if err != nil {
			return rec.ProcessResponse{}, fmt.Errorf("combineResults failed : %v", err)
		}
		userWeight := getUserWeight(input, res)
		product := recogConf * configWeight * userWeight // intermediate weight
		if res.Ok {
			anyOk = true
		} else {
			product = 0.0
		}
		converted := rec.SubProcessResponse{
			Ok:                res.Ok,
			RecognitionResult: res.RecognitionResult,
			RecordingID:       res.RecordingID,
			Message:           res.Message,
			InputConfidence: map[string]float64{
				"recogniser": recogConf,
				"config":     configWeight,
				"user":       userWeight,
				"product":    product,
			},
			Confidence: product, // intermediate value
		}
		convertedResults = append(convertedResults, converted)
		totalConf += roundConfidence(converted.Confidence)
		log.Printf("CombineResults:1 [%s] '%s' | conf=%f cw=%f uw=%f => %f", res.Source(), res.RecognitionResult, recogConf, configWeight, userWeight, product)
	}
	// re-compute conf relative to the sum of weights
	var totalConfs = make(map[string]float64) // result string => sum of confidence measures for responses with this result
	for i, res := range convertedResults {
		newConf := 0.0
		if totalConf > 0 {
			newConf = roundConfidence(res.Confidence / totalConf)
		}
		res.Confidence = newConf
		convertedResults[i] = res // update the slice with new value
		totalConfs[res.RecognitionResult] = totalConfs[res.RecognitionResult] + res.Confidence
		log.Printf("CombineResults:2 [%s] '%s' | conf=%f", res.Source(), res.RecognitionResult, res.Confidence)
	}

	if resErr != nil {
		return rec.ProcessResponse{}, nil
	}
	if !anyOk {
		return rec.ProcessResponse{}, fmt.Errorf("no results available")
	}

	var selected rec.ProcessResponse
	if len(convertedResults) > 0 {
		bestGuess, weight := getBestGuess(totalConfs)
		selected = rec.ProcessResponse{Ok: true,
			RecordingID:       input.RecordingID,
			Message:           "selected result",
			RecognitionResult: bestGuess,
			Confidence:        weight}
	} else {
		selected = rec.ProcessResponse{Ok: false,
			RecordingID:       input.RecordingID,
			Message:           "No result from server",
			RecognitionResult: ""}
	}
	if includeOriginalResponses {
		selected.ComponentResults = convertedResults
	}
	return selected, nil
}
