package rec

import (
	"bytes"
	"encoding/json"
)

func PrettyMarshal(thing interface{}) ([]byte, error) {
	var res []byte

	j, err := json.Marshal(thing)
	if err != nil {
		return res, err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, j, "", "\t")
	if err != nil {
		return res, err
	}
	res = prettyJSON.Bytes()
	return res, nil
}
