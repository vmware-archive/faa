package slackcommand

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Request struct {
	Command string
	Message string
}

type Delegate interface {
	Handle(Request) error
}

type Server struct {
	VerificationToken string
	Delegate          Delegate
}

type ErrResponse struct {
	Message string `json:"error"`
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrResponse{err.Error()})
		return
	}

	t := r.PostFormValue("token")
	if t != s.VerificationToken {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrResponse{"token verification failed"})
		return
	}

	text := r.PostFormValue("text")
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrResponse{"missing text"})
		return
	}

	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrResponse{"must provide command and text"})
		return
	}

	slackRequest := Request{
		Command: parts[0],
		Message: parts[1],
	}

	err = s.Delegate.Handle(slackRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrResponse{err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(nil)
}
