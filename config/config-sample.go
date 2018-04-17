package config

var ConfigSample = Config{
	AudioDir:   "audio_dir",
	ServerPort: 9993,
	Recognisers: []Recogniser{
		Recogniser{
			Name: "default",
			Type: KaldiGStreamer,
			Weights: map[string]float64{
				"word":    0.7,
				"char":    0.7,
				"default": 0.7,
			},
			Cmd: "http://192.168.0.105:8080/client/dynamic/recognize",
		},
		Recogniser{
			Name:     "elexia_20180412",
			Type:     Tensorflow,
			Disabled: true,
			Weights: map[string]float64{
				"word":    0.7,
				"char":    0.0,
				"default": 0.7,
			},
			Cmd: "bash /home/hanna/go/src/github.com/stts-se/audioproc/tensorflow/scripts/classify.sh /home/hanna/progz/tensorflow/simple_audio_recognition/e-lexia-20180412/my_frozen_graph.pb /home/hanna/progz/tensorflow/simple_audio_recognition/e-lexia-20180412/speech_commands_train/conv_labels.txt {wavfile}",
		},
		Recogniser{
			Name:     "nst-aeneas-4",
			Type:     Tensorflow,
			Disabled: true,
			Weights: map[string]float64{
				"default": 0.7,
				"word":    0.7,
				"char":    0.0,
			},
			Cmd: "bash /home/hanna/go/src/github.com/stts-se/audioproc/tensorflow/scripts/classify.sh /home/hanna/progz/tensorflow/simple_audio_recognition/danish-test-aeneas-4/my_frozen_graph.pb /home/hanna/progz/tensorflow/simple_audio_recognition/danish-test-aeneas-4/speech_commands_train/conv_labels.txt {wavfile}",
		},
		Recogniser{
			Name: "nst_10000_adapted_182/8000",
			Type: PocketSphinx,
			Weights: map[string]float64{
				"word":    0.0,
				"char":    0.7,
				"default": 0.7,
			},
			Cmd: "http://localhost:8000/rec?audio_file={wavfile}",
		},
		Recogniser{
			Name: "elexia_448/9999",
			Type: PocketSphinx,
			Weights: map[string]float64{
				"word":    0.7,
				"char":    0.0,
				"default": 0.7,
			},
			Cmd: "http://localhost:9999/rec?audio_file={wavfile}",
		},
	},
}
