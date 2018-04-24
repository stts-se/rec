package rec

import (
	"fmt"
	"testing"
	//"github.com/sergi/go-diff/diffmatchpatch"
)

func Test_PR_PrettyJSON(t *testing.T) {

	input := ProcessResponse{
		Ok:                true,
		Confidence:        0.9999,
		RecognitionResult: "bi",
		RecordingID:       "tmprecid0",
		Message:           "4 out of 4 recognisers responded",
		ComponentResults: []RecogniserResponse{
			RecogniserResponse{
				Status: true,
				InputConfidence: map[string]float64{
					"config":     1,
					"product":    0.77671,
					"recogniser": 0.77671,
					"user":       1,
				},
				Confidence:        0.6321,
				RecognitionResult: "bi",
				RecordingID:       "tmprecid0",
				Message:           "",
				Source:            "tensorflow_cmd|elexia_503_20180412",
			},
			RecogniserResponse{
				Status: true,
				InputConfidence: map[string]float64{
					"config":     0,
					"product":    0,
					"recogniser": 0.30963,
					"user":       1,
				},
				Confidence:        0,
				RecognitionResult: "o",
				RecordingID:       "tmprecid0",
				Message:           "",
				Source:            "tensorflow_cmd|nst_chars_20170410",
			},
		},
	}

	expect := `{
	"ok": true,
	"confidence": 0.9999,
	"recognition_result": "bi",
	"recording_id": "tmprecid0",
	"message": "4 out of 4 recognisers responded",
	"component_results": [
		{
			"status": true,
			"input_confidence": {"config": 1, "product": 0.77671, "recogniser": 0.77671, "user": 1},
			"confidence": 0.6321,
			"recognition_result": "bi",
			"recording_id": "tmprecid0",
			"message": "",
			"source": "tensorflow_cmd|elexia_503_20180412"
		},
		{
			"status": true,
			"input_confidence": {"config": 0, "product": 0, "recogniser": 0.30963, "user": 1},
			"confidence": 0,
			"recognition_result": "o",
			"recording_id": "tmprecid0",
			"message": "",
			"source": "tensorflow_cmd|nst_chars_20170410"
		}
	]
}`

	res, err := input.PrettyJSON()
	if err != nil {
		t.Errorf("%v", err)
	} else {
		if res != expect {
			t.Errorf("EXPECTED \n<<%s>>\n, FOUND: \n<<%s>>\n", expect, res)
			fmt.Println()
		}
	}
}
