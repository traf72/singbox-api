package api

import (
	"log"
	"net/http"

	"github.com/traf72/singbox-api/internal/api/header"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/utils"
)

var jsonSerializeOptions = &utils.JSONOptions{Indent: "    ", EscapeHTML: false}

func SendJson(w http.ResponseWriter, body any) {
	header.SetContentType(w, header.ContentTypeJson)

	if err := utils.ToJSON(w, body, jsonSerializeOptions); err != nil {
		SendInternalServerError(w, apperr.NewFatalErr("JsonEncodingError", err.Error()))
	}
}

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
