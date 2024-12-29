package api

import (
	"log"
	"net/http"

	"github.com/traf72/singbox-api/internal/apperr"
)

func SendBadRequest(w http.ResponseWriter, err string) {
	http.Error(w, err, http.StatusBadRequest)
}

func SendNotFound(w http.ResponseWriter, err string) {
	http.Error(w, err, http.StatusNotFound)
}

func SendConflict(w http.ResponseWriter, err string) {
	http.Error(w, err, http.StatusConflict)
}

func SendInternalServerError(w http.ResponseWriter, err apperr.Err) {
	log.Printf("%d %s: %s", http.StatusInternalServerError, err.Code(), err.Msg())
	http.Error(w, "", http.StatusInternalServerError)
}

func SendError(w http.ResponseWriter, e apperr.Err) {
	switch e.Kind() {
	case apperr.Validation:
		SendBadRequest(w, e.Msg())
	case apperr.NotFound:
		SendNotFound(w, e.Msg())
	case apperr.Conflict:
		SendConflict(w, e.Msg())
	default:
		SendInternalServerError(w, e)
	}
}
