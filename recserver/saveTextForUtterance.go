package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func saveTextForUtterance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	scriptName := vars["scriptname"]
	userName := vars["username"]
	recID := vars["utteranceid"]
	text := vars["text"]
	fmt.Printf("SN: %s\tUN: %s\tID: %s\tText: %s\n", scriptName, userName, recID, text)

	fmt.Fprintf(w, "OK!")
}
