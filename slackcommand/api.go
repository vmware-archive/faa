package slackcommand

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/schema"
)

type Command struct {
	Token       string `schema:"token"`
	TeamID      string `schema:"team_id"`
	TeamDomain  string `schema:"team_domain"`
	ChannelID   string `schema:"channel_id"`
	ChannelName string `schema:"channel_name"`
	UserID      string `schema:"user_id"`
	UserName    string `schema:"user_name"`
	Command     string `schema:"command"`
	Text        string `schema:"text"`
	ResponseURL string `schema:"response_url"`
}

type Delegate interface {
	Handle(Command) (string, error)
}

type Server struct {
	VerificationToken string
	Delegate          Delegate
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

	var c Command

	err = schema.NewDecoder().Decode(&c, r.PostForm)
	if err != nil {
		json.NewEncoder(w).Encode(NewErrResponse(err.Error()))
		return
	}

	if c.Token != s.VerificationToken {
		json.NewEncoder(w).Encode(NewErrResponse("token verification failed"))
		return
	}

	msg, err := s.Delegate.Handle(c)
	if err != nil {
		json.NewEncoder(w).Encode(NewErrResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(NewOKResponse(msg))
}
