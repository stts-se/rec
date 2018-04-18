package config

var ConfigSample = Config{
	AudioDir:   "audio_dir",
	ServerPort: 9993,
	Recognisers: []Recogniser{
		Recogniser{
			Name: "english",
			Type: KaldiGStreamer,
			Weights: map[string]float32{
				"_other_": 1.0,
				"default": 0.0,
			},
			Cmd: "http://192.168.0.105:8080/client/dynamic/recognize",
		},
		Recogniser{
			Name:     "elexia_20180412",
			Type:     Tensorflow,
			Disabled: true,
			Weights: map[string]float32{
				"default": 0.99,
			},
			Cmd: "bash /home/hanna/go/src/github.com/stts-se/audioproc/tensorflow/scripts/classify.sh /home/hanna/progz/tensorflow/simple_audio_recognition/e-lexia-20180412/my_frozen_graph.pb /home/hanna/progz/tensorflow/simple_audio_recognition/e-lexia-20180412/speech_commands_train/conv_labels.txt {wavfile}",
		},
		Recogniser{
			Name:     "nst-aeneas-4",
			Type:     Tensorflow,
			Disabled: true,
			Weights: map[string]float32{
				"char":    0.6,
				"default": 0.0,
			},
			Cmd: "bash /home/hanna/go/src/github.com/stts-se/audioproc/tensorflow/scripts/classify.sh /home/hanna/progz/tensorflow/simple_audio_recognition/danish-test-aeneas-4/my_frozen_graph.pb /home/hanna/progz/tensorflow/simple_audio_recognition/danish-test-aeneas-4/speech_commands_train/conv_labels.txt {wavfile}",
		},
		Recogniser{
			Name: "nst_10000_adapted_182/8000",
			Type: PocketSphinx,
			Weights: map[string]float32{
				"char":    0.0,
				"default": 0.8,
			},
			Cmd: "http://localhost:8000/rec?audio_file={wavfile}",
		},
		Recogniser{
			Name: "elexia_448/9999",
			Type: PocketSphinx,
			Weights: map[string]float32{
				"default": 0.8,
			},
			Cmd: "http://localhost:9999/rec?audio_file={wavfile}",
		},
	},
}
