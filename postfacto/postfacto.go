package postfacto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

type RetroClient struct {
	Host string
	ID   string
}

type Category string

const (
	CategoryHappy Category = "happy"
	CategoryMeh   Category = "meh"
	CategorySad   Category = "sad"
)

type RetroItem struct {
	Description string   `json:"description"`
	Category    Category `json:"category"`
}

func (c *RetroClient) Add(i RetroItem) error {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(i)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/retros/%s/items", c.Host, c.ID), b)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		b, _ := httputil.DumpResponse(res, true)
		return fmt.Errorf("unexpected response code (%d) - %s", res.StatusCode, string(b))
	}

	return nil
}
