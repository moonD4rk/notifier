package bark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Provider struct {
	Server string
	Key    string
	url    string
}

const (
	DefaultBarkServer = "api.day.app"
)

func New(key, server string) *Provider {
	if key == "" {
		return nil
	}
	p := &Provider{
		Server: server,
		Key:    key,
	}
	if server != "" {
		p.url = fmt.Sprintf("https://%s/push", server)
	} else {
		p.url = fmt.Sprintf("https://%s/push", DefaultBarkServer)
	}
	return p
}

func (p *Provider) Send(subject, content string) error {
	type postData struct {
		DeviceKey string `json:"device_key"`
		Title     string `json:"title"`
		Body      string `json:"body,omitempty"`
		Badge     int    `json:"badge,omitempty"`
		Sound     string `json:"sound,omitempty"`
		Icon      string `json:"icon,omitempty"`
		Group     string `json:"group,omitempty"`
		URL       string `json:"url,omitempty"`
	}
	pd := &postData{
		DeviceKey: p.Key,
		Title:     subject,
		Body:      content,
		Sound:     "alarm.caf",
	}
	data, err := json.Marshal(pd)
	if err != nil {
		return err
	}

	resp, err := http.Post(p.url, "application/json; charset=utf-8", bytes.NewReader(data))
	if err != nil {
		return err
	}
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("statusCode: %d, body: %v", resp.StatusCode, string(result))
	}
	return nil
}
