package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

var MyConfig Config

type Config struct {
	AudioDir       string            `json:"audio_dir"`
	ServerPort     string            `json:"server_port"`
	PocketSphinx   map[string]string `json:"pocketsphinx"`
	TensorFlow     map[string]string `json:"tensorflow"`
	KaldiGStreamer map[string]string `json:"kaldi_gstreamer"`
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
	return res, nil

}

func (cfg Config) ListRecognizers() []string {
	res := []string{}
	for name, _ := range cfg.PocketSphinx {
		res = append(res, "pocketsphinx|"+name)
	}
	for name, _ := range cfg.TensorFlow {
		res = append(res, "tensorflow|"+name)
	}
	for name, _ := range cfg.KaldiGStreamer {
		res = append(res, "gstreamer kaldi|"+name)
	}
	return res
}
