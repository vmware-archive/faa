package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"fmt"
	"strings"

	"github.com/concourse/faa/postfacto"
	"github.com/concourse/faa/slackcommand"
)

func main() {
	var (
		port string
		ok   bool
	)
	port, ok = os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	vToken, ok := os.LookupEnv("SLACK_VERIFICATION_TOKEN")
	if !ok {
		panic(errors.New("Must provide SLACK_VERIFICATION_TOKEN"))
	}

	retroID, ok := os.LookupEnv("POSTFACTO_RETRO_ID")
	if !ok {
		panic(errors.New("Must provide POSTFACTO_RETRO_ID"))
	}

	techRetroID, ok := os.LookupEnv("POSTFACTO_TECH_RETRO_ID")
	if !ok {
		panic(errors.New("Must provide POSTFACTO_TECH_RETRO_ID"))
	}

	c := &postfacto.RetroClient{
		Host: "https://retro-api.cfapps.io",
		ID:   retroID,
	}

	t := &postfacto.RetroClient{
		Host: "https://retro-api.cfapps.io",
		ID:   techRetroID,
	}

	server := slackcommand.Server{
		VerificationToken: vToken,
		Delegate: &PostfactoSlackDelegate{
			RetroClient:     c,
			TechRetroClient: t,
		},
	}

	http.Handle("/", server)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type PostfactoSlackDelegate struct {
	RetroClient     *postfacto.RetroClient
	TechRetroClient *postfacto.RetroClient
}

type Command string

const (
	CommandHappy Command = "happy"
	CommandMeh   Command = "meh"
	CommandSad   Command = "sad"
	CommandTech  Command = "tech"
)

func (d *PostfactoSlackDelegate) Handle(r slackcommand.Command) (string, error) {
	parts := strings.SplitN(r.Text, " ", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("must be in the form of '%s [happy/meh/sad/tech] [message]'", r.Command)
	}

	c := parts[0]
	description := parts[1]

	var (
		client   *postfacto.RetroClient
		category postfacto.Category
	)

	switch Command(c) {
	case CommandHappy:
		category = postfacto.CategoryHappy
		client = d.RetroClient
	case CommandMeh:
		category = postfacto.CategoryMeh
		client = d.RetroClient
	case CommandSad:
		category = postfacto.CategorySad
		client = d.RetroClient
	case CommandTech:
		category = postfacto.CategoryHappy
		client = d.TechRetroClient
	default:
		return "", errors.New("unknown command: must provide one of 'happy', 'meh', 'sad', or 'tech'")
	}

	retroItem := postfacto.RetroItem{
		Category:    category,
		Description: fmt.Sprintf("%s [%s]", description, r.UserName),
	}

	err := client.Add(retroItem)
	if err != nil {
		return "", err
	}

	return "retro item added", nil
}
