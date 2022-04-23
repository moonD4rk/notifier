package dingtalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/moond4rk/notifier/internal/encrypter"
)

type provider struct {
	Token  string `yaml:"token,omitempty"`
	Secret string `yaml:"secret,omitempty"`
}

func New(token, secret string) *provider {
	if token == "" {
		return nil
	}
	return &provider{
		Token:  token,
		Secret: secret,
	}
}

func (p *provider) Send(subject, content string) error {
	data, err := buildPostData(subject, content)
	if err != nil {
		return errors.Wrap(err, "failed to create message")
	}
	url, err := encrypter.DingTalkURL(p.Token, p.Secret)
	if err != nil {
		return errors.Wrap(err, "build dingtalk url error")
	}
	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewReader(data))
	if err != nil {
		return errors.Wrap(err, "send dingtalk request failed")
	}
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("statusCode: %d, body: %v", resp.StatusCode, string(result))
		return errors.Wrap(err, "dingtalk message response error")
	}
	return errors.Wrap(err, "send dingtalk message failed")
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
