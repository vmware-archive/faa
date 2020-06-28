package postfacto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type RetroClient struct {
	Host     string
	ID       string
	Password string
	token    string
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
	WriteKey    string   `json:"writeKey"`
}

func (c *RetroClient) Login() error {
	var m = make(map[string]map[string]string)

	m["retro"] = make(map[string]string)
	m["retro"]["password"] = c.Password

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/retros/%s/login", c.Host, c.ID), b)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		b, _ := httputil.DumpResponse(res, true)
		return fmt.Errorf("unexpected response code (%d) - %s", res.StatusCode, string(b))
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(body, &objmap)
	if err != nil {
		return err
	}

	var p string
	err = json.Unmarshal(*objmap["token"], &p)
	if err != nil {
		return err
	}

	c.token = p
	return nil
}

func (c *RetroClient) Add(i RetroItem) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(i)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/retros/%s/items", c.Host, c.ID), b)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	switch res.StatusCode {
	case http.StatusCreated:
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized. try setting *_RETRO_PASSWORD var(s)")
	default:
		b, _ := httputil.DumpResponse(res, true)
		return fmt.Errorf("unexpected response code (%d) - %s", res.StatusCode, string(b))
	}

	return nil
}
