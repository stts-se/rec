package aggregator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/config"
)

func round(f float64) string {
	rounded := roundConfidence(f) // defined in aggregator.go
	return fmt.Sprintf("%.4f", rounded)
}

func testEqualConf(exp string, got float64) bool {
	return exp == round(got)
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
	expW := round(0.625)
	msg = fmt.Sprintf("expected output weight %s", expW)
	result, err = CombineResults(cfg, pInput, input, true)
	if err != nil {
		t.Errorf("%s, got error : %v", msg, err)
	} else if !testEqualConf(expW, result.Confidence) {
		t.Errorf("%s, got %s", msg, round(result.Confidence))
	}

	if result.RecognitionResult != "bix" {
		t.Errorf("expected %s, got %s", "bi", result.RecognitionResult)
	}

	for _, resp := range result.ComponentResults {
		if resp.Source == "kaldigstreamer|rc1" {
			expW = round(.375)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc2" {
			expW = round(.25)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc3" {
			expW = round(.375)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round(resp.Confidence))
			}
		} else {
			t.Errorf("unknown recogniser name: %s", resp.Source)
		}
	}

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
	expW := round(0.8333)
	msg = fmt.Sprintf("expected output weight %s", expW)
	result, err = CombineResults(cfg, pInput, input, true)
	if err != nil {
		t.Errorf("%s, got error : %v", msg, err)
	} else if !testEqualConf(expW, result.Confidence) {
		t.Errorf("%s, got %s", msg, round(result.Confidence))
	}

	if result.RecognitionResult != "bi" {
		t.Errorf("expected %s, got %s", "bi", result.RecognitionResult)
	}

	for _, resp := range result.ComponentResults {
		if resp.Source == "kaldigstreamer|rc1" {
			expW = round(.5)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc2" {
			expW = round(.3333)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round(resp.Confidence))
			}
		} else if resp.Source == "kaldigstreamer|rc3" {
			expW = round(.1667)
			if !testEqualConf(expW, resp.Confidence) {
				msg = fmt.Sprintf("expected output weight %s", expW)
				t.Errorf("%s, got %s", msg, round(resp.Confidence))
			}
		} else {
			t.Errorf("unknown recogniser name: %s", resp.Source)
		}
	}

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
	expW := round(0.9999)
	msg = fmt.Sprintf("expected output weight %s", expW)
	result, err = CombineResults(cfg, pInput, input, true)
	if err != nil {
		t.Errorf("%s, got error : %v", msg, err)
	} else if !testEqualConf(expW, result.Confidence) {
		t.Errorf("%s, got %s", msg, round(result.Confidence))
	}

	if result.RecognitionResult != "ba" {
		t.Errorf("expected %s, got %s", "bi", result.RecognitionResult)
	}

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
