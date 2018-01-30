package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

// TO DO remove:
func dummyInstantiateUtts() {

	uttLists.currentUttForUser = make(map[string]int)
	uttLists.uttsForUser = make(map[string][]utt)

	utts1 := []utt{
		{"rec_0001", "This is utterance number one"},
		{"rec_0002", "Utterance number two."},
		{"rec_0003", "Well, number three"},
	}
	utts2 := []utt{
		{"rec_0001", "This is utterance number one, user two"},
		{"rec_0002", "Utterance number two, user two."},
		{"rec_0003", "Well, number three, user two"},
	}
	uttLists.uttsForUser["user0001"] = utts1
	uttLists.uttsForUser["user0002"] = utts2
}

func init() {
	dummyInstantiateUtts()
}

type utt struct {
	uttID string
	text  string
}

type utteranceLists struct {
	sync.Mutex
	currentUttForUser map[string]int
	uttsForUser       map[string][]utt
}

var uttLists = utteranceLists{}

type nextUttResponse struct {
	UserName    string `json:"username"`
	Text        string `json:"text"`
	RecordingID string `json:"recording_id"`
	Message     string `json:"message"`
}

func getUttRelativeToCurrent(userName string, uttIndex int) (utt, string) {
	var res utt
	var msg string

	uttLists.Lock()
	defer uttLists.Unlock()

	var newIndex int

	var utterances []utt

	if utts, ok := uttLists.uttsForUser[userName]; !ok {
		msg := fmt.Sprintf("get_next_utterance: no utterances for user '%s'", userName)
		log.Print(msg)
		return res, msg
	} else {

		if len(uttLists.uttsForUser[userName]) == 0 {
			msg := fmt.Sprintf("get_next_utterance: no utterances for user '%s'", userName)
			log.Print(msg)
			return res, msg
		}

		utterances = utts

		// Not first utterace
		if currIndex, ok := uttLists.currentUttForUser[userName]; ok {
			newIndex = uttIndex + currIndex
		} else { //first utterance in list, currIndex == 0
			newIndex = 0 //uttIndex + currIndex
		}
	}

	if newIndex < 0 {
		newIndex = 0
		msg = "at first utterance"
	}
	if newIndex > len(utterances)-1 {
		newIndex = len(utterances) - 1
		msg = "at last utterance"
	}

	uttLists.currentUttForUser[userName] = newIndex
	return utterances[newIndex], msg
}

func getNextUtterance(w http.ResponseWriter, r *http.Request) {
	var res nextUttResponse

	vars := mux.Vars(r)
	userName := strings.ToLower(vars["username"])

	utt, msg := getUttRelativeToCurrent(userName, 1)

	res.UserName = userName
	res.RecordingID = utt.uttID
	res.Text = utt.text
	res.Message = msg

	resJSON, err := json.Marshal(res)
	if err != nil {
		msg0 := fmt.Sprintf("get_next_utterance: failed JSON conversion of struct : %v", err)
		log.Print(msg0)
		http.Error(w, msg0, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(resJSON))
}

func getPreviousUtterance(w http.ResponseWriter, r *http.Request) {
	var res nextUttResponse

	vars := mux.Vars(r)
	userName := strings.ToLower(vars["username"])

	utt, msg := getUttRelativeToCurrent(userName, -1)

	res.UserName = userName
	res.RecordingID = utt.uttID
	res.Text = utt.text
	res.Message = msg

	resJSON, err := json.Marshal(res)
	if err != nil {
		msg0 := fmt.Sprintf("get_previous_utterance: failed JSON conversion of struct : %v", err)
		log.Print(msg0)
		http.Error(w, msg0, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(resJSON))
}
