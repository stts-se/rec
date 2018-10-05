package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	"github.com/stts-se/rec"
)

// TO DO remove:
//func dummyInstantiateUtts() {

//uttLists.currentUttForUser = make(map[string]int)
//uttLists.uttsForUser = make(map[string][]utterance)

//utts1 := []utterance{
//	{RecordingID: "rec_0001", Text: "This is utterance number one"},
//	{RecordingID: "rec_0002", Text: "Utterance number two."},
//	{RecordingID: "rec_0003", Text: "Well, number three"},
//}
//utts2 := []utterance{
//	{RecordingID: "rec_0001", Text: "This is utterance number one, user two"},
//	{RecordingID: "rec_0002", Text: "Utterance number two, user two."},
//	{RecordingID: "rec_0003", Text: "Well, number three, user two"},
//}
//uttLists.uttsForUser["user0001"] = utts1
//uttLists.uttsForUser["user0002"] = utts2
//}

// TO DO remove:
//func init() {
//	loadUtteranceLists(config.MyConfig.AudioDir)
//	dummyInstantiateUtts()
//}

// type utt struct {
// 	uttID string
// 	text  string
// }

type utteranceLists struct {
	sync.Mutex
	currentUttForUser map[string]int
	uttsForUser       map[string][]rec.Utterance
}

func newUtteranceLists() utteranceLists {
	return utteranceLists{
		currentUttForUser: make(map[string]int),
		uttsForUser:       make(map[string][]rec.Utterance),
	}
}

var uttLists = newUtteranceLists() //utteranceLists{}

// type utterance struct {
// 	UserName    string `json:"username"`
// 	Text        string `json:"text"`
// 	RecordingID string `json:"recording_id"`
// 	Message     string `json:"message"`
// 	Num         int    `json:"num"`
// 	Of          int    `json:"of"`
// }

func getUtterance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userName := strings.ToLower(vars["scriptname"])
	uttIndex, err := strconv.Atoi(vars["uttindex"])
	if err != nil {
		msg := fmt.Sprintf("getUtterance: failed to convert argument into integer: %v", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	uttLists.Lock()
	defer uttLists.Unlock()

	var utts []rec.Utterance

	utts, ok := uttLists.uttsForUser[userName]
	if !ok || len(utts) == 0 {
		msg := fmt.Sprintf("get_next_utterance: no utterances for user '%s'", userName)
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest)

		return
	}

	//utterances = utts

	if uttIndex <= 0 || uttIndex > len(utts) {
		msg := fmt.Sprintf("no utterance number %d", uttIndex)
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest)

		return
	}

	res := utts[uttIndex-1]
	res.Of = len(utts)
	res.Num = uttIndex
	// Since we changed the dir structure to
	// <audir_dir>/script_dir/user_dir, we cannot associate a
	// username to an utterance of a script
	//res.UserName = userName

	resJSON, err := json.Marshal(res)
	if err != nil {
		msg0 := fmt.Sprintf("get_next_utterance: failed JSON conversion of struct : %v", err)
		log.Print(msg0)
		http.Error(w, msg0, http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(resJSON))

}

func getUttRelativeToCurrent(userName string, uttIndex int) (rec.Utterance, error) {
	var res rec.Utterance
	var msg string

	uttLists.Lock()
	defer uttLists.Unlock()

	var newIndex int

	var utterances []rec.Utterance

	utts, ok := uttLists.uttsForUser[userName]
	if !ok || len(utts) == 0 {
		msg := fmt.Sprintf("get_next_utterance: no utterances for user '%s'", userName)
		log.Print(msg)
		return res, fmt.Errorf(msg)
	}
	//else
	utterances = utts

	// Not first utterace
	if currIndex, ok := uttLists.currentUttForUser[userName]; ok {
		newIndex = uttIndex + currIndex
	} else { //first utterance in list, currIndex == 0
		newIndex = 0 //uttIndex + currIndex
	}
	//}

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

// TODO adds data to global var uttLists
// TODO contents of different .utt files are collapsed. Want to keep them apart?
func loadUtteranceLists(dirPath string) /*(utteranceLists,*/ error {
	//var res = newUtteranceLists()

	files, err := filepath.Glob(filepath.Join(dirPath, "*", "*.utt"))
	if err != nil {
		return fmt.Errorf("loadUtteranceLists: failed to list user *.utt files : %v", err)
	}

	uttLists.Lock()
	defer uttLists.Unlock()
	for _, f := range files {
		//fmt.Printf("FN: %s\n", f)

		//fmt.Printf("base: %s\n", base)
		// Parent dir name of file is "user name"
		userName := path.Base(path.Dir(f))
		utts, err := readUttFile(f)
		if err != nil {
			return fmt.Errorf("loadUtteranceLists: failed to read file : %v", err)
		}

		uttLists.uttsForUser[userName] = append(uttLists.uttsForUser[userName], utts...)
	}

	return nil
}

func readUttFile(fn string) ([]rec.Utterance, error) {
	var res []rec.Utterance

	bytes, err := ioutil.ReadFile(fn)
	if err != nil {
		return res, fmt.Errorf("readUttFile: %v", err)
	}

	lines := strings.Split(string(bytes), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}

		// TODO validate line
		fs := strings.SplitN(l, "\t", 2)
		if len(fs) != 2 || len(fs[0]) == 0 || len(fs[1]) == 0 {
			log.Printf("readUttFile: skipping line of '%s': %s", fn, l)
			continue
		}
		u := rec.Utterance{RecordingID: fs[0], Text: fs[1]}
		//fmt.Printf("%#v\n", u)
		res = append(res, u)

	}

	return res, nil
}
