package dingtalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/moond4rk/notifier/internal/crypto"
)

type Provider struct {
	Token  string `yaml:"token,omitempty"`
	Secret string `yaml:"secret,omitempty"`
}

func New(token, secret string) *Provider {
	if token == "" {
		return nil
	}
	return &Provider{
		Token:  token,
		Secret: secret,
	}
}

func (p *Provider) Send(subject, content string) error {
	data, err := buildPostData(subject, content)
	if err != nil {
		return fmt.Errorf("failed to create message %w", err)
	}
	link, err := crypto.DingTalkURL(p.Token, p.Secret)
	if err != nil {
		return fmt.Errorf("build dingtalk link error %w", err)
	}
	resp, err := http.Post(link, "application/json; charset=utf-8", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("send dingtalk request failed %w", err)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer func() {
		// Close the body and check for errors
		if cerr := resp.Body.Close(); cerr != nil {
			// Handle the error, log it, etc. Here we're just logging.
			log.Printf("failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("statusCode: %d, body: %v", resp.StatusCode, string(result))
		return fmt.Errorf("dingtalk message response error %w", err)
	}
	return nil
}

func buildPostData(subject, content string) ([]byte, error) {
	content = fmt.Sprintf("### %s\n>%s", subject, content)
	type postData struct {
		MsgType  string `json:"msgtype"`
		Markdown struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		} `json:"markdown"`
	}
	pd := &postData{MsgType: "markdown"}
	pd.Markdown.Title = subject
	pd.Markdown.Text = content
	data, err := json.Marshal(pd)
	if err != nil {
		return nil, err
	}
	return data, err
}
