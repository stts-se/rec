package aggregator

import (
	"github.com/stts-se/rec/config"
)

var ConfigTest1 = config.Config{
	AudioDir:   "audio_dir",
	ServerPort: 9993,
	Recognisers: []config.Recogniser{
		config.Recogniser{
			Name: "english",
			Type: "kaldigstreamer",
			Cmd:  "http://192.168.0.105:8080/client/dynamic/recognize",
			Weights: map[string]float64{
				"output:_other_":   1.0,
				"output:_silence_": 1.0,
				"default":          0.0,
			},
		},
		config.Recogniser{
			Name: "elexia_20180412",
			Type: "tensorflow",
			Cmd:  "bash /home/hanna/go/src/github.com/stts-se/audioproc/tensorflow/scripts/classify.sh /home/hanna/progz/tensorflow/simple_audio_recognition/e-lexia-20180412/my_frozen_graph.pb /home/hanna/progz/tensorflow/simple_audio_recognition/e-lexia-20180412/speech_commands_train/conv_labels.txt {wavfile}",
			Weights: map[string]float64{
				"output:_silence_": 1.0,
			},
		},
		config.Recogniser{
			Name: "nst-aeneas-4",
			Type: "tensorflow",
			Cmd:  "bash /home/hanna/go/src/github.com/stts-se/audioproc/tensorflow/scripts/classify.sh /home/hanna/progz/tensorflow/simple_audio_recognition/danish-test-aeneas-4/my_frozen_graph.pb /home/hanna/progz/tensorflow/simple_audio_recognition/danish-test-aeneas-4/speech_commands_train/conv_labels.txt {wavfile}",
			Weights: map[string]float64{
				"input:_char_":     1.0,
				"output:_silence_": 1.0,
				"default":          0.0,
			},
		},
		config.Recogniser{
			Name: "nst_10000_adapted_467",
			Type: "pocketsphinx",
			Cmd:  "http://localhost:8000/rec?audio_file={wavfile}",
			Weights: map[string]float64{
				"input:_char_":     0.0,
				"output:_silence_": 0.25,
				"default":          0.5,
			},
		},
		config.Recogniser{
			Name: "nst_46912",
			Type: "pocketsphinx",
			Cmd:  "http://localhost:9090/rec?audio_file={wavfile}",
			Weights: map[string]float64{
				"input:_char_":     0.0,
				"output:_silence_": 0.25,
				"default":          0.5,
			},
		},
		config.Recogniser{
			Name: "nst_46912_adapted_467",
			Type: "pocketsphinx",
			Cmd:  "http://localhost:9091/rec?audio_file={wavfile}",
			Weights: map[string]float64{
				"input:_char_":     0.0,
				"output:_silence_": 0.5,
			},
		},
		config.Recogniser{
			Name: "elexia_448",
			Type: "pocketsphinx",
			Cmd:  "http://localhost:9999/rec?audio_file={wavfile}",
			Weights: map[string]float64{
				"output:_silence_": 0.5,
			},
		},
	},
}
