package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

var MyConfig Config

var Tensorflow = "tensorflow"
var PocketSphinx = "pocketsphinx"
var KaldiGStreamer = "kaldigstreamer"

type Recogniser struct {
	Name     string             `json:"name"`
	Type     string             `json:"type"`
	Cmd      string             `json:"cmd"`
	Weights  map[string]float64 `json:"weights,omitempty"`
	Disabled bool               `json:"disabled"`
}
type Config struct {
	AudioDir    string       `json:"audio_dir"`
	ServerPort  int32        `json:"server_port"`
	Recognisers []Recogniser `json:"recognisers,omitempty"`
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
	log.Printf("Loaded config: %#v\n", res)
	return res, nil

}

func (cfg Config) String() string {
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

// LongName returns cfg.Type <vertical bar> cfg.Name, e.g. pocketsphinx|elexia_448/9999
func (rec Recogniser) LongName() string {
	return fmt.Sprintf("%s|%s", rec.Type, rec.Name)
}

func (cfg Config) RecogniserNames() []string {
	res := []string{}
	for _, rec := range cfg.Recognisers {
		res = append(res, rec.LongName())
	}
	return res
}
