package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/stts-se/rec/admin"
)

func listUsers(w http.ResponseWriter, r *http.Request) {
	// audioDir global var in recserver.go
	users, err := admin.ListUsers(audioDir)
	if err != nil {
		log.Printf("failed to list users : %v", err)
		http.Error(w, "filed to list users", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", strings.Join(users, "\n"))
}

// TODO Validate/sanitize input
func addUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userName := vars["username"]
	userName = strings.ToLower(userName)
	// audioDir global var in recserver.go
	err := admin.AddUser(audioDir, userName)
	if err != nil {
		msg := fmt.Sprintf("failed add user '%s' : %v", userName, err)
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest) // internal server error?
		return
	}

	fmt.Fprintf(w, "added user '"+userName+"'\n")
}

// TODO Remove this call?
func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userName := vars["username"]
	userName = strings.ToLower(userName)
	// audioDir global var in recserver.go
	err := admin.DeleteUser(audioDir, userName)
	if err != nil {
		msg := fmt.Sprintf("failed to delete user '%s' : %v", userName, err)
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest) // internal server error?
		return
	}
	fmt.Fprintf(w, "deleted user '"+userName+"'\n")
}

func getUtts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userName := vars["username"]
	userName = strings.ToLower(userName)
	uttLists, err := admin.ListUtts(audioDir, userName)
	if err != nil {
		msg := fmt.Sprintf("filed to list utterance for '%s' : %v", userName, err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	uttListsJSON, err := json.Marshal(uttLists)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%v\n", string(uttListsJSON))
}
