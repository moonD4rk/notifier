package serverchan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

var defaultHost = "sctapi.ftqq.com"

func (p *Provider) Send(subject, content string) error {
	link := fmt.Sprintf("https://%s/%s.send", defaultHost, p.SendKey)
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
	resp, err := http.Post(link, "application/json; charset=utf-8", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("http post %w", err)
	}
	defer func() {
		// Close the body and check for errors
		if cerr := resp.Body.Close(); cerr != nil {
			// Handle the error, log it, etc. Here we're just logging.
			log.Printf("failed to close response body: %v", cerr)
		}
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf("send server chan failed %w", fmt.Errorf("http status code: %d", resp.StatusCode))
	}
	return nil
}
