package Repeater

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Message string `json:"message"`
}

func Send(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(status)
	body, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func SendOK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func SendError(w http.ResponseWriter, status int, err string) {
	Send(w, status, Error{Message: err})
}
