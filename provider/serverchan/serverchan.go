package serverchan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type provider struct {
	UserID  string
	SendKey string
}

func New(userID, sendKey string) *provider {
	if sendKey == "" {
		return nil
	}
	return &provider{
		UserID:  userID,
		SendKey: sendKey,
	}
}

func (p *provider) Send(subject, content string) error {
	var (
		defaultServer = "sctapi.ftqq.com"
	)
	url := fmt.Sprintf("https://%s/%s.send", defaultServer, p.SendKey)
	type postData struct {
		Text string `json:"text"`
		Desp string `json:"desp"`
	}
	pd := &postData{
		Text: subject,
		Desp: content,
	}
	data, err := json.Marshal(pd)
	if err != nil {
		return errors.Wrap(err, "json marshal")
	}
	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewReader(data))
	if err != nil {
		return errors.Wrap(err, "http post")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err := fmt.Errorf("http status code: %d", resp.StatusCode)
		return errors.Wrap(err, "send server chan failed")
	}
	return nil
}
