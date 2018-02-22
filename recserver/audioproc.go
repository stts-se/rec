package main

import (
	"fmt"
	"net/http"
)

func audioProc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", "audioProc: not implemented")
}
