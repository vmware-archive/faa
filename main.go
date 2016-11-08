package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"strconv"

	"fmt"
	"strings"

	"github.com/chendrix/faa/postfacto"
	"github.com/chendrix/faa/slackcommand"
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

	retroIDS, ok := os.LookupEnv("POSTFACTO_RETRO_ID")
	if !ok {
		panic(errors.New("Must provide POSTFACTO_RETRO_ID"))
	}
	retroID, err := strconv.Atoi(retroIDS)
	if err != nil {
		panic(errors.New("POSTFACTO_RETRO_ID must be an integer"))
	}

	c := &postfacto.RetroClient{
		Host: "https://retro-api.cfapps.io",
		ID:   retroID,
	}

	server := slackcommand.Server{
		VerificationToken: vToken,
		Delegate: &PostfactoSlackDelegate{
			RetroClient: c,
		},
	}

	http.Handle("/", server)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type PostfactoSlackDelegate struct {
	RetroClient *postfacto.RetroClient
}

func (d *PostfactoSlackDelegate) Handle(r slackcommand.Command) (string, error) {
	parts := strings.SplitN(r.Text, " ", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("must be in the form of '%s [happy/meh/sad] [message]'", r.Command)
	}

	c := parts[0]
	description := parts[1]

	var category postfacto.Category
	switch postfacto.Category(c) {
	case postfacto.CategoryHappy:
		category = postfacto.CategoryHappy
	case postfacto.CategoryMeh:
		category = postfacto.CategoryMeh
	case postfacto.CategorySad:
		category = postfacto.CategorySad
	default:
		return "", errors.New("unknown postfacto category: must provide one of 'happy', 'meh', or 'sad'")
	}

	retroItem := postfacto.RetroItem{
		Category:    category,
		Description: fmt.Sprintf("%s [%s]", description, r.UserName),
	}

	err := d.RetroClient.Add(retroItem)
	if err != nil {
		return "", err
	}

	return "retro item added", nil
}
