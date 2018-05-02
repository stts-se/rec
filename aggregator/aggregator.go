package aggregator

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

var vowels = map[string]bool{
	"e": true, "y": true, "u": true, "i": true, "o": true, "å": true, "a": true, "ö": true, "ä": true, "ø": true, "æ": true,
}
var conses = map[string]bool{
	"q": true, "w": true, "r": true, "t": true, "p": true, "s": true, "d": true, "f": true, "g": true, "h": true, "j": true, "k": true, "l": true, "z": true, "x": true, "c": true, "v": true, "b": true, "n": true, "m": true,
}

var voiced = map[string]bool{
	"r": true, "d": true, "g": true, "h": true, "j": true, "l": true, "v": true, "b": true, "n": true, "m": true,
}

var devoiced = map[string]bool{
	"q": true, "w": true, "t": true, "p": true, "s": true, "f": true, "k": true, "z": true, "x": true, "c": true,
}

func isVowel(s string) bool {
	_, ok := vowels[s]
	return ok
}
func isCons(s string) bool {
	_, ok := conses[s]
	return ok
}

func isVoiced(s string) bool {
	if isVowel(s) {
		return true
	}
	_, ok := voiced[s]
	return ok
}
func isDevoiced(s string) bool {
	if isVowel(s) {
		return false
	}
	_, ok := devoiced[s]
	return ok
}

func isChar(s string) bool {
	return len([]rune(s)) == 1
}

func isWord(s string) bool {
	return len([]rune(s)) > 1
}

var knownPropertyRE = regexp.MustCompile("^_(word|char|vowel|cons|voiced|devoiced)_$")

func isKnownProperty(s string) bool {
	return knownPropertyRE.MatchString(s)
}

func isProperty(s string) bool {
	return strings.HasPrefix(s, "_") && strings.HasSuffix(s, "_")
}

func roundConfidence(fl float64) float64 {
	unit := 0.0001
	return math.Round(fl/unit) * unit
}

func getUserWeight(input rec.ProcessInput, res rec.RecogniserResponse) float64 {
	rcName := res.Source
	if w, ok := input.Weights[rcName]; ok {
		return w
	}
	return 1.0
}

func getConfigWeight(input rec.ProcessInput, res rec.RecogniserResponse, recName2Weights map[string]config.Recogniser) (float64, string, error) {
	rc, ok := recName2Weights[res.Source]
	if !ok {
		msg := fmt.Sprintf("no recogniser defined for %s", res.Source)
		return 0.0, "", fmt.Errorf("%s", msg)
	}
	ws := rc.Weights
	if w, ok := ws["input:"+input.Text]; ok {
		return w, "input:" + input.Text, nil
	} else if w, ok := ws["output:"+res.RecognitionResult]; ok {
		return w, "output:" + res.RecognitionResult, nil
	} else if w, ok := ws["input:_char_"]; ok && isChar(input.Text) {
		return w, "input:_char_", nil
	} else if w, ok := ws["input:_word_"]; ok && isWord(input.Text) {
		return w, "input:_word_", nil
	} else if w, ok := ws["output:_char_"]; ok && isChar(res.RecognitionResult) {
		return w, "output:_char_", nil
	} else if w, ok := ws["output:_word_"]; ok && isWord(res.RecognitionResult) {
		return w, "output:_word_", nil
	} else if w, ok := ws["default"]; ok {
		return w, "default", nil
	}
	return 1.0, "", nil
}

const usePropertyConfs = true

func getBestGuess(totalConfs map[string]float64) (string, float64) {
	var bestConf = -1.0
	var bestGuess string
	for guess, conf := range totalConfs {
		if conf > bestConf && (!usePropertyConfs || !isKnownProperty(guess)) {
			bestConf = conf
			bestGuess = guess
		}
	}
	if bestConf >= 1.0 { // we can never be 100% sure
		bestConf = roundConfidence(0.9999)
	}
	return bestGuess, bestConf
}

func updatePropertyConfs(totalConfs map[string]float64, propertyConfs map[string]float64) map[string]float64 {
	updated := totalConfs
	// recalculate conf for properties
	for res, conf := range totalConfs {
		if isVowel(res) {
			updated[res] = conf + propertyConfs["_vowel_"]
			updated["_vowel_"] = totalConfs["_vowel_"] + conf
		}
		if isCons(res) {
			updated[res] = conf + propertyConfs["_cons_"]
			updated["_cons_"] = totalConfs["_cons_"] + conf
		}
		if isChar(res) {
			updated[res] = conf + propertyConfs["_char_"]
			updated["_char_"] = totalConfs["_char_"] + conf
		}
		if isWord(res) {
			updated[res] = conf + propertyConfs["_word_"]
			updated["_word_"] = totalConfs["_word_"] + conf
		}
		if isVoiced(res) {
			updated[res] = conf + propertyConfs["_voiced_"]
			updated["_voiced_"] = totalConfs["_voiced_"] + conf
		}
		if isDevoiced(res) {
			updated[res] = conf + propertyConfs["_devoiced_"]
			updated["_devoiced_"] = totalConfs["_devoiced_"] + conf
		}
	}
	return updated
}

func CombineResults(cfg config.Config, input rec.ProcessInput, inputResults []rec.RecogniserResponse, includeOriginalResponses bool) (rec.ProcessResponse, error) {
	var resErr error
	var convertedResults = inputResults
	var recName2Weights = make(map[string]config.Recogniser)
	var recNames = make(map[string]bool)
	for _, rc := range cfg.EnabledRecognisers() {
		name := rc.LongName()
		recNames[name] = true
		if strings.TrimSpace(name) == "" {
			return rec.ProcessResponse{}, fmt.Errorf("empty recogniser name in source %v", rc)
		}
		if !rc.Disabled {
			recName2Weights[name] = rc
		}
	}

	// compute initial weights (recogniser conf * config defined conf * user defined conf)
	var totalConf = 0.0
	var nProcessSuccess = 0
	var sourceMap = make(map[string]bool)
	for i, res := range convertedResults {
		if res.Source == "" {
			return rec.ProcessResponse{}, fmt.Errorf("empty source in response : %v", res)
		}
		if _, ok := recNames[res.Source]; !ok {
			return rec.ProcessResponse{}, fmt.Errorf("no recogniser defined for source %s", res.Source)
		}
		if _, ok := sourceMap[res.Source]; ok {
			return rec.ProcessResponse{}, fmt.Errorf("recogniser sources must be unique, found repeated source name %s in %v", res.Source, res)
		}
		sourceMap[res.Source] = true

		if res.Status == true {
			nProcessSuccess += 1
		}
		recogConf := res.Confidence // input confidence from recogniser
		if recogConf < 0.0 {        // below zero => unknown/undefined => default value 1.0
			recogConf = 1.0
		}
		configWeight, configLog, err := getConfigWeight(input, res, recName2Weights)
		if err != nil {
			return rec.ProcessResponse{}, fmt.Errorf("CombineResults failed : %v", err)
		}
		configName := "config"
		if configLog != "" {
			configName = "config|" + configLog
		}
		userWeight := getUserWeight(input, res)
		combined := recogConf * configWeight * userWeight // intermediate weight
		res.InputConfidence = map[string]float64{
			"recogniser": recogConf,
			configName:   configWeight,
			"user":       userWeight,
			"combined":   combined,
		}
		res.Confidence = combined // the intermediate confidence value
		convertedResults[i] = res // update the slice with new values
		totalConf += roundConfidence(res.Confidence)
	}
	// re-compute conf relative to the sum of weights
	var totalConfs = make(map[string]float64) // result string => sum of confidence measures for responses with this result
	// ... and save property confs for later use
	var propertyConfs = make(map[string]float64)
	for i, res := range convertedResults {
		newConf := 0.0
		if totalConf > 0 {
			newConf = roundConfidence(res.Confidence / totalConf)
		}
		res.Confidence = newConf
		convertedResults[i] = res // update the slice with new value
		totalConfs[res.RecognitionResult] = totalConfs[res.RecognitionResult] + res.Confidence
		if isKnownProperty(res.RecognitionResult) {
			propertyConfs[res.RecognitionResult] = propertyConfs[res.RecognitionResult] + res.Confidence
			// } else if isProperty(res.RecognitionResult) {
			// 	return rec.ProcessResponse{}, fmt.Errorf("unknown property : %s", res.RecognitionResult)
		}
		log.Printf("CombineResults %v", res)
	}
	if usePropertyConfs {
		totalConfs = updatePropertyConfs(totalConfs, propertyConfs)
	}
	if resErr != nil {
		return rec.ProcessResponse{}, nil
	}

	var selected rec.ProcessResponse
	var message = fmt.Sprintf("%d out of %d recognisers responded", nProcessSuccess, len(convertedResults))
	if len(convertedResults) > 0 {
		bestGuess, weight := getBestGuess(totalConfs)
		selected = rec.ProcessResponse{
			// Ok:  bestGuess == input.Text
			Ok:                true, // 20170423: always 'true' for backward compatibility, TODO: better value/better field
			RecordingID:       input.RecordingID,
			Message:           message,
			RecognitionResult: bestGuess,
			Confidence:        weight,
		}
	} else {
		selected = rec.ProcessResponse{
			Ok:                false,
			RecordingID:       input.RecordingID,
			Message:           fmt.Sprintf("no result from server; %s", message),
			RecognitionResult: "",
		}
	}
	if includeOriginalResponses {
		selected.ComponentResults = convertedResults
	}
	return selected, nil
}
