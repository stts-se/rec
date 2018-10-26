package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

// TODO This might be a temporary way to handle manually edited
// versions of the text corresponding to an utterance.
func saveTextForUtterance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	scriptName := vars["scriptname"]
	userName := vars["username"]
	recID := vars["utteranceid"]
	text := vars["text"]
	fmt.Printf("SN: %s\tUN: %s\tID: %s\tText: %s\n", scriptName, userName, recID, text)

	// audioDir defined in recserver.go
	fn := path.Join(audioDir, scriptName, userName, recID+".txt")

	err := ioutil.WriteFile(fn, []byte(text+"\n"), 0644)
	if err != nil {
		log.Printf("saveTextForUtterance: failed to write file '%s' : %v", fn, err)
		msg := fmt.Sprintf("failed to save text for utterance %s/%s/%s", scriptName, userName, recID)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	//fmt.Printf("SN: %s\tUN: %s\tID: %s\tText: %s\tPath: %s\n", scriptName, userName, recID, text, fn)

	fmt.Fprintf(w, "Saved text for recording %s", recID)
}
