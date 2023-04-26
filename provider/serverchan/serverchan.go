package serverchan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Provider struct {
	UserID  string
	SendKey string
}

func New(userID, sendKey string) *Provider {
	if sendKey == "" {
		return nil
	}
	return &Provider{
		UserID:  userID,
		SendKey: sendKey,
	}
}

func (p *Provider) Send(subject, content string) error {
	defaultServer := "sctapi.ftqq.com"
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
		return fmt.Errorf("json marshal %w", err)
	}
	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("http post %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("send server chan failed %w", fmt.Errorf("http status code: %d", resp.StatusCode))
	}
	return nil
}
