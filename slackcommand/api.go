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
	Handle(Request) (string, error)
}

type Server struct {
	VerificationToken string
	Delegate          Delegate
}

type ErrResponse struct {
	Message string `json:"error"`
}

type Response struct {
	Type string `json:"response_type"`
	Text string `json:"text"`
}

func NewOKResponse(text string) Response {
	return Response{
		Type: "in_channel",
		Text: text,
	}
}

func NewErrResponse(text string) Response {
	return Response{
		Type: "ephemeral",
		Text: text,
	}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := r.ParseForm()
	if err != nil {
		json.NewEncoder(w).Encode(NewErrResponse(err.Error()))
		return
	}

	t := r.PostFormValue("token")
	if t != s.VerificationToken {
		json.NewEncoder(w).Encode(NewErrResponse("token verification failed"))
		return
	}

	text := r.PostFormValue("text")
	if text == "" {
		json.NewEncoder(w).Encode(NewErrResponse("missing text"))
		return
	}

	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		json.NewEncoder(w).Encode(NewErrResponse("must provide command and text"))
		return
	}

	slackRequest := Request{
		Command: parts[0],
		Message: parts[1],
	}

	msg, err := s.Delegate.Handle(slackRequest)
	if err != nil {
		json.NewEncoder(w).Encode(NewErrResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(NewOKResponse(msg))
}
