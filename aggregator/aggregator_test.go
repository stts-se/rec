package aggregator

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

func round2S(f float64) string {
	rounded := roundConfidence(f) // defined in aggregator.go
	return fmt.Sprintf("%.4f", rounded)
}

func testEqualConf(exp string, got float64) bool {
	return exp == round2S(got)
}

func testCheckSum(t *testing.T, result rec.ProcessResponse) {
	var sum = 0.0
	for _, rc := range result.ComponentResults {
		sum += rc.Confidence
	}
	unit := 0.01
	rounded := math.Round(sum/unit) * unit
	if rounded != 1.00 {
		t.Errorf("Expected sum of confidences to be near 1.0000, but found '%.2f': %v", rounded, result.ComponentResults)
	}
}

func Test_CombineResults_Computations2_UserAndRecogniserWeights1(t *testing.T) {
	var recID = "test_0001"
	var pInput = rec.ProcessInput{
		UserName:    "tmpuser",
		Audio:       rec.Audio{},
		Text:        "i",
		RecordingID: recID,
		Weights: map[string]float64{
			"kaldigstreamer|rc3": 3,
		},
	}
	var cfg = config.Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []config.Recogniser{
			config.Recogniser{
				Name: "rc1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
				Weights: map[string]float64{
					"output:_word_": 0.6,
					"default":       0.8,
				},
			},
			config.Recogniser{
				Name: "rc2",
				Type: "kaldigstreamer",
				Weights: map[string]float64{
					"default": 1.5,
				},
				Cmd: "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc3",
				Type: "kaldigstreamer",
				Weights: map[string]float64{
					"input:_char_": 0.0,
				},
				Cmd: "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}
	var input []rec.RecogniserResponse
	var result rec.ProcessResponse
	var err error
	var msg string

	input = []rec.RecogniserResponse{
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1.5, // weighted => 1.5 * 0.6 * 1.0 = 0.9 | => 0.375
			RecognitionResult: "bi",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc1",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1, // weighted => 1.0 * 1.5 * 1.0 = 1.5 | => .625
			RecognitionResult: "bix",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc2",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0.5, // weighted => 0.5 * 0.0 * 3.0 = 0.0 | => 0.0
			RecognitionResult: "bi",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc3",
		},
	}
	expW := round2S(0.625)
	msg = fmt.Sprintf("expected output weight %s", expW)
	result, err = CombineResults(cfg, pInput, input, true)
	if err != nil {
		t.Errorf("%s, got error : %v", msg, err)
	} else if !testEqualConf(expW, result.Confidence) {
		t.Errorf("%s, got %s", msg, round2S(result.Confidence))
	}

	if result.RecognitionResult != "bix" {
		t.Errorf("expected %s, got %s", "bix", result.RecognitionResult)
	}

	for _, resp := range result.ComponentResults {
		if resp.Source == "kaldigstreamer|rc1" {
			expW = round2S(.375)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc2" {
			expW = round2S(.625)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc3" {
			expW = round2S(.0)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else {
			t.Errorf("unknown recogniser name: %s", resp.Source)
		}
	}
	testCheckSum(t, result)
}

func Test_CombineResults_Computations2_UserAndRecogniserWeights2(t *testing.T) {
	var recID = "test_0001"
	var pInput = rec.ProcessInput{
		UserName:    "tmpuser",
		Audio:       rec.Audio{},
		Text:        "bi",
		RecordingID: recID,
		Weights: map[string]float64{
			"kaldigstreamer|rc2": 3,
		},
	}
	var cfg = config.Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []config.Recogniser{
			config.Recogniser{
				Name: "rc1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
				Weights: map[string]float64{
					"output:_word_": 0.6,
					"default":       0.8,
				},
			},
			config.Recogniser{
				Name: "rc2",
				Type: "kaldigstreamer",
				Weights: map[string]float64{
					"output:bi": 2.0,
					"default":   1.5,
				},
				Cmd: "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc3",
				Type: "kaldigstreamer",
				Weights: map[string]float64{
					"input:_char_": 0.0,
					"default":      0.3,
				},
				Cmd: "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}
	var input []rec.RecogniserResponse
	var result rec.ProcessResponse
	var err error
	var msg string

	input = []rec.RecogniserResponse{
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1.5, // weighted => 1.5 * 0.8 * 1.0 = 1.2 / 7.35 => 0.1633
			RecognitionResult: "i",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc1",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1, // weighted => 1.0 * 2.0 * 3.0 = 6.0 / 7.35 => 0.8163
			RecognitionResult: "bi",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc2",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0.5, // weighted => 0.5 * 0.3 * 1.0 = 0.15 / 7.35 => 0.0204
			RecognitionResult: "bi",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc3",
		},
	}
	expW := round2S(0.8163 + .0204)
	msg = fmt.Sprintf("expected output weight %s", expW)
	result, err = CombineResults(cfg, pInput, input, true)
	if err != nil {
		t.Errorf("%s, got error : %v", msg, err)
	} else if !testEqualConf(expW, result.Confidence) {
		t.Errorf("%s, got %s", msg, round2S(result.Confidence))
	}

	if result.RecognitionResult != "bi" {
		t.Errorf("expected %s, got %s", "bi", result.RecognitionResult)
	}

	for _, resp := range result.ComponentResults {
		if resp.Source == "kaldigstreamer|rc1" {
			expW = round2S(.1633)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc2" {
			expW = round2S(.8163)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc3" {
			expW = round2S(.0204)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else {
			t.Errorf("unknown recogniser name: %s", resp.Source)
		}
	}
	testCheckSum(t, result)
}

func Test_CombineResults_Computations2_UserWeights(t *testing.T) {
	var recID = "test_0001"
	var pInput = rec.ProcessInput{
		UserName:    "tmpuser",
		Audio:       rec.Audio{},
		Text:        "",
		RecordingID: recID,
		Weights: map[string]float64{
			"kaldigstreamer|rc3": 3,
		},
	}
	var cfg = config.Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []config.Recogniser{
			config.Recogniser{
				Name: "rc1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc2",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc3",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}
	var input []rec.RecogniserResponse
	var result rec.ProcessResponse
	var err error
	var msg string

	input = []rec.RecogniserResponse{
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1.5, // weighted => 1.5*1 / 4 = .375
			RecognitionResult: "bix",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc1",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1, // weighted => 1*1 / 4 = .25
			RecognitionResult: "bix",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc2",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0.5, // weighted => 0.5 * 3 / 4 .375
			RecognitionResult: "bin",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc3",
		},
	}
	expW := round2S(0.625)
	msg = fmt.Sprintf("expected output weight %s", expW)
	result, err = CombineResults(cfg, pInput, input, true)
	if err != nil {
		t.Errorf("%s, got error : %v", msg, err)
	} else if !testEqualConf(expW, result.Confidence) {
		t.Errorf("%s, got %s", msg, round2S(result.Confidence))
	}

	if result.RecognitionResult != "bix" {
		t.Errorf("expected %s, got %s", "bix", result.RecognitionResult)
	}

	for _, resp := range result.ComponentResults {
		if resp.Source == "kaldigstreamer|rc1" {
			expW = round2S(.375)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc2" {
			expW = round2S(.25)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc3" {
			expW = round2S(.375)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else {
			t.Errorf("unknown recogniser name: %s", resp.Source)
		}
	}
	testCheckSum(t, result)
}

func Test_CombineResults_Computations1(t *testing.T) {
	var recID = "test_0001"
	var pInput = rec.ProcessInput{
		UserName:    "tmpuser",
		Audio:       rec.Audio{},
		Text:        "",
		RecordingID: recID,
		Weights: map[string]float64{
			"tensorflow|usermodifiable": 1.3,
		},
	}
	var cfg = config.Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []config.Recogniser{
			config.Recogniser{
				Name: "rc1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc2",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc3",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}
	var input []rec.RecogniserResponse
	var result rec.ProcessResponse
	var err error
	var msg string

	input = []rec.RecogniserResponse{
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1.5, // weighted => .5
			RecognitionResult: "bi",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc1",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1, // weighted => .3333
			RecognitionResult: "bi",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc2",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0.5, // weighted => .1667
			RecognitionResult: "bin",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc3",
		},
	}
	expW := round2S(0.8333)
	msg = fmt.Sprintf("expected output weight %s", expW)
	result, err = CombineResults(cfg, pInput, input, true)
	if err != nil {
		t.Errorf("%s, got error : %v", msg, err)
	} else if !testEqualConf(expW, result.Confidence) {
		t.Errorf("%s, got %s", msg, round2S(result.Confidence))
	}

	if result.RecognitionResult != "bi" {
		t.Errorf("expected %s, got %s", "bi", result.RecognitionResult)
	}

	for _, resp := range result.ComponentResults {
		if resp.Source == "kaldigstreamer|rc1" {
			expW = round2S(.5)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc2" {
			expW = round2S(.3333)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc3" {
			expW = round2S(.1667)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round2S(resp.Confidence))
			}
		} else {
			t.Errorf("unknown recogniser name: %s", resp.Source)
		}
	}
	testCheckSum(t, result)
}

func Test_CombineResults_AlgorithmAlwaysBelowOne(t *testing.T) {
	var recID = "test_0001"
	var pInput = rec.ProcessInput{
		UserName:    "tmpuser",
		Audio:       rec.Audio{},
		Text:        "",
		RecordingID: recID,
		Weights: map[string]float64{
			"tensorflow|usermodifiable": 1.3,
		},
	}
	var cfg = config.Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []config.Recogniser{
			config.Recogniser{
				Name: "rc1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc2",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}
	var input []rec.RecogniserResponse
	var result rec.ProcessResponse
	var err error
	var msg string

	input = []rec.RecogniserResponse{
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1,
			RecognitionResult: "ba",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc1",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        1,
			RecognitionResult: "ba",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc2",
		},
	}
	expW := round2S(0.9999)
	msg = fmt.Sprintf("expected output weight %s", expW)
	result, err = CombineResults(cfg, pInput, input, true)
	if err != nil {
		t.Errorf("%s, got error : %v", msg, err)
	} else if !testEqualConf(expW, result.Confidence) {
		t.Errorf("%s, got %s", msg, round2S(result.Confidence))
	}

	if result.RecognitionResult != "ba" {
		t.Errorf("expected %s, got %s", "ba", result.RecognitionResult)
	}
	testCheckSum(t, result)
}

func Test_CombineResults_EmptyName(t *testing.T) {
	var recID = "test_0001"
	var pInput = rec.ProcessInput{
		UserName:    "tmpuser",
		Audio:       rec.Audio{},
		Text:        "",
		RecordingID: recID,
		Weights: map[string]float64{
			"tensorflow|usermodifiable": 1.3,
		},
	}
	var cfg = config.Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []config.Recogniser{
			config.Recogniser{
				Name: "rc1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc2",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}
	var input []rec.RecogniserResponse
	var err error
	var msg string

	input = []rec.RecogniserResponse{
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0.6321,
			RecognitionResult: "bi",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc1",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0,
			RecognitionResult: "o",
			RecordingID:       recID,
			Message:           "",
			Source:            "",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0,
			RecognitionResult: "o",
			RecordingID:       recID,
			Message:           "",
			Source:            "rc2",
		},
	}
	msg = "expected error for empty name"
	_, err = CombineResults(cfg, pInput, input, true)
	if err == nil {
		t.Errorf("%s in %v", msg, input)
	} else if !strings.Contains(err.Error(), "empty source") {
		t.Errorf("%s in %v, got : %v", msg, input, err)
	}
}

func Test_CombineResults_RepeatedNames(t *testing.T) {
	var recID = "test_0001"
	var pInput = rec.ProcessInput{
		UserName:    "tmpuser",
		Audio:       rec.Audio{},
		Text:        "",
		RecordingID: recID,
		Weights: map[string]float64{
			"tensorflow|usermodifiable": 1.3,
		},
	}
	var cfg = config.Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []config.Recogniser{
			config.Recogniser{
				Name: "rc1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc2",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}
	var input []rec.RecogniserResponse
	var err error
	var msg string

	input = []rec.RecogniserResponse{
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0.6321,
			RecognitionResult: "bi",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc1",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0,
			RecognitionResult: "o",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc1",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0,
			RecognitionResult: "o",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc2",
		},
	}
	msg = "expected error for repeated source names"
	_, err = CombineResults(cfg, pInput, input, true)
	if err == nil {
		t.Errorf("%s in %v", msg, input)
	} else if !strings.Contains(err.Error(), "repeated source") {
		t.Errorf("%s in %v, got : %v", msg, input, err)
	}
}

func Test_CombineResults_UndefinedSources(t *testing.T) {
	var recID = "test_0001"
	var pInput = rec.ProcessInput{
		UserName:    "tmpuser",
		Audio:       rec.Audio{},
		Text:        "",
		RecordingID: recID,
		Weights: map[string]float64{
			"tensorflow|usermodifiable": 1.3,
		},
	}
	var cfg = config.Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []config.Recogniser{
			config.Recogniser{
				Name: "rc1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			config.Recogniser{
				Name: "rc2",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}
	var input []rec.RecogniserResponse
	var err error
	var msg string

	input = []rec.RecogniserResponse{
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0.6321,
			RecognitionResult: "bi",
			RecordingID:       recID,
			Message:           "",
			Source:            "kaldigstreamer|rc1",
		},
		rec.RecogniserResponse{
			Status:            true,
			Confidence:        0,
			RecognitionResult: "o",
			RecordingID:       recID,
			Message:           "",
			Source:            "rc2",
		},
	}
	msg = "expected error for undefined source"
	_, err = CombineResults(cfg, pInput, input, true)
	if err == nil {
		t.Errorf("%s in %v", msg, input)
	} else if !strings.Contains(err.Error(), "no recogniser defined for") {
		t.Errorf("%s in %v, got : %v", msg, input, err)
	}
}
