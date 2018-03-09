package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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
