package api

import (
	"net/http"
)

func SendBadRequest(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusBadRequest)
}
