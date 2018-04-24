package config

import (
	//"fmt"
	"strings"
	"testing"
)

func Test_Config1(t *testing.T) {

	var config Config
	var err error
	var msg string

	// EMPTY RECOGNISER NAME - NOT OK
	config = Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []Recogniser{
			Recogniser{
				Name: "",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			Recogniser{
				Name: "test2",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}

	msg = "expected error for empty name"
	err = config.test()
	if err == nil {
		t.Errorf("%s in %v", msg, config)
	} else if !strings.Contains(err.Error(), "empty recogniser name") {
		t.Errorf("%s in %v, got : %v", msg, config, err)
	}

	// REPEATED RECOGNISER NAMES FOR THE SAME TYPE - NOT OK
	config = Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []Recogniser{
			Recogniser{
				Name: "test1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			Recogniser{
				Name: "test1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}

	msg = "expected error for repeated recogniser name"
	err = config.test()
	if err == nil {
		t.Errorf("%s in %v", msg, config)
	} else if !strings.Contains(err.Error(), "recogniser names must be unique") {
		t.Errorf("%s in %v, got : %v", msg, config, err)
	}

	// REPEATED RECOGNISER NAMES FOR DIFFERENT TYPES - OK
	config = Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []Recogniser{
			Recogniser{
				Name: "test1",
				Type: "kaldigstreamer",
			},
			Recogniser{
				Name: "test1",
				Type: "tensorflow",
			},
		},
	}

	msg = "didn't expect error for repeated recogniser name"
	err = config.test()
	if err != nil {
		t.Errorf("%s in %v", msg, config)
	}

	// INVALID RECOGNISER TYPE - NOT OK
	config = Config{
		AudioDir:              "audio_dir",
		ServerPort:            9993,
		FailOnRecogniserError: true,
		Recognisers: []Recogniser{
			Recogniser{
				Name: "test1",
				Type: "kaldigstreamer",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
			Recogniser{
				Name: "test1",
				Type: "anytypeiwant",
				Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			},
		},
	}

	msg = "expected error for invalid recogniser type"
	err = config.test()
	if err == nil {
		t.Errorf("%s in %v", msg, config)
	} else if !strings.Contains(err.Error(), "invalid recogniser type") {
		t.Errorf("%s in %v, got : %v", msg, config, err)
	}

}
