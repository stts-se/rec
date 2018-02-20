package main

import (
	"net/http"
)

func animDemo(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../recclient/animationdemo/index.html")
}
