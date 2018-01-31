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
	uttLists.uttsForUser = make(map[string][]utterance)

	utts1 := []utterance{
		{RecordingID: "rec_0001", Text: "This is utterance number one"},
		{RecordingID: "rec_0002", Text: "Utterance number two."},
		{RecordingID: "rec_0003", Text: "Well, number three"},
	}
	utts2 := []utterance{
		{RecordingID: "rec_0001", Text: "This is utterance number one, user two"},
		{RecordingID: "rec_0002", Text: "Utterance number two, user two."},
		{RecordingID: "rec_0003", Text: "Well, number three, user two"},
	}
	uttLists.uttsForUser["user0001"] = utts1
	uttLists.uttsForUser["user0002"] = utts2
}

// TO DO remove:
func init() {
	dummyInstantiateUtts()
}

// type utt struct {
// 	uttID string
// 	text  string
// }

type utteranceLists struct {
	sync.Mutex
	currentUttForUser map[string]int
	uttsForUser       map[string][]utterance
}

var uttLists = utteranceLists{}

type utterance struct {
	UserName    string `json:"username"`
	Text        string `json:"text"`
	RecordingID string `json:"recording_id"`
	Message     string `json:"message"`
	Num         int    `json:"num"`
	Of          int    `json:"of"`
}

func getUttRelativeToCurrent(userName string, uttIndex int) (utterance, error) {
	var res utterance
	var msg string

	uttLists.Lock()
	defer uttLists.Unlock()

	var newIndex int

	var utterances []utterance

	if utts, ok := uttLists.uttsForUser[userName]; !ok {
		msg := fmt.Sprintf("get_next_utterance: no utterances for user '%s'", userName)
		log.Print(msg)
		return res, fmt.Errorf(msg)
	} else {

		if len(uttLists.uttsForUser[userName]) == 0 {
			msg := fmt.Sprintf("get_next_utterance: no utterances for user '%s'", userName)
			log.Print(msg)
			return res, fmt.Errorf(msg)
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

	utterances[newIndex].UserName = userName
	utterances[newIndex].Message = msg
	utterances[newIndex].Num = newIndex + 1 // Number, not index
	utterances[newIndex].Of = len(utterances)
	return utterances[newIndex], nil
}

func getNextUtterance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userName := strings.ToLower(vars["username"])

	res, err := getUttRelativeToCurrent(userName, 1)
	if err != nil {
		msg0 := fmt.Sprintf("get_next_utterance: failed getting utterance : %v", err)
		log.Print(msg0)
		http.Error(w, msg0, http.StatusInternalServerError)
		return
	}

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
	vars := mux.Vars(r)
	userName := strings.ToLower(vars["username"])

	res, err := getUttRelativeToCurrent(userName, -1)
	if err != nil {
		msg0 := fmt.Sprintf("get_previous_utterance: failed getting utterance : %v", err)
		log.Print(msg0)
		http.Error(w, msg0, http.StatusInternalServerError)
		return
	}

	res.UserName = userName

	resJSON, err := json.Marshal(res)
	if err != nil {
		msg0 := fmt.Sprintf("get_previous_utterance: failed JSON conversion of struct : %v", err)
		log.Print(msg0)
		http.Error(w, msg0, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(resJSON))
}
