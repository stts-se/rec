package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

var MyConfig Config

var Tensorflow = "tensorflow"
var TensorflowCmd = "tensorflow_cmd"
var PocketSphinx = "pocketsphinx"
var PocketSphinxWithFilter = "pocketsphinx_withfilter"
var KaldiGStreamer = "kaldigstreamer"

type Recogniser struct {
	Name     string             `json:"name"`
	Type     string             `json:"type"`
	Cmd      string             `json:"cmd"`
	Weights  map[string]float64 `json:"weights,omitempty"`
	Disabled bool               `json:"disabled,omitempty"`
}
type Config struct {
	AudioDir              string       `json:"audio_dir"`
	ServerPort            int          `json:"server_port"`
	FailOnRecogniserError bool         `json:"fail_on_recogniser_error"`
	Recognisers           []Recogniser `json:"recognisers,omitempty"`
}

func (cfg Config) test() error {
	if strings.TrimSpace(cfg.AudioDir) == "" {
		return fmt.Errorf("empty audio_dir in config %v", cfg)
	}
	recMap := make(map[string]Recogniser)
	for _, rc := range cfg.Recognisers {
		if strings.TrimSpace(rc.Name) == "" {
			return fmt.Errorf("empty recogniser name in config %v", cfg)
		}
		switch rc.Type {
		case Tensorflow:
		case TensorflowCmd:
		case KaldiGStreamer:
		case PocketSphinx:
		case PocketSphinxWithFilter:
		default:
			return fmt.Errorf("invalid recogniser type: %s", rc.Type)
		}
		if _, ok := recMap[rc.LongName()]; ok {
			return fmt.Errorf("recogniser names must be unique, found repeated name %s", rc.LongName())
		}
		recMap[rc.LongName()] = rc
	}
	return nil
}

func NewConfig(filePath string) (Config, error) {
	log.Printf("Loading config file: %s", filePath)
	bts, err := ioutil.ReadFile(filePath)
	res := Config{}
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("failed to read config file : %v", err)
	}
	err = json.Unmarshal(bts, &res)
	if err != nil {
		log.Printf("failure : %v\n", err)
		return res, fmt.Errorf("failed to unmarshal : %v", err)
	}
	err = res.test()
	if err != nil {
		return res, fmt.Errorf("init tests failed for config file %s : %v", filePath, err)
	}
	return res, nil
}

func (cfg Config) EnabledRecognisers() []Recogniser {
	res := []Recogniser{}
	for _, rc := range cfg.Recognisers {
		if !rc.Disabled {
			res = append(res, rc)
		}
	}
	return res
}

func (cfg Config) PrettyString() string {
	bts, err := json.Marshal(cfg)
	if err != nil {
		log.Printf("failed to process JSON : %v\n", err)
		return fmt.Sprintf("JSON NOT AVAILABLE FOR CONFIG %#v", cfg)
	}
	var prettyBody bytes.Buffer
	err = json.Indent(&prettyBody, bts, "", "\t")
	if err != nil {
		log.Printf("failed to process JSON : %v\n", err)
		return fmt.Sprintf("JSON NOT AVAILABLE FOR CONFIG %#v", cfg)
	}
	return string(prettyBody.Bytes())
}

// LongName returns cfg.Type <vertical bar> cfg.Name, e.g. pocketsphinx|nst_chars
func (rec Recogniser) LongName() string {
	return fmt.Sprintf("%s|%s", rec.Type, rec.Name)
}

func (cfg Config) RecogniserNames() []string {
	res := []string{}
	for _, rc := range cfg.Recognisers {
		res = append(res, rc.Name)
	}
	return res
}
