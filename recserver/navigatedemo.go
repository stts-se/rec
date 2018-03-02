package main

import (
	"net/http"
)

func navigateDemo(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../recclient/navigatedemo/index.html")
}
