package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

var MyConfig Config

type Config struct {
	KaldiGStreamerURL string `json:"kaldi_gstreamer_url"`
	AudioDir          string `json:"audio_dir"`
	ServerPort        string `json:"server_port"`
	TensorflowCmd     string `json:"tensorflow_cmd"`
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
