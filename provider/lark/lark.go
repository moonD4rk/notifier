package lark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/moond4rk/notifier/internal/crypto"
)

type Provider struct {
	Token  string
	Secret string
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
	data, err := crypto.LarkData(subject, content, p.Secret)
	url := fmt.Sprintf("https://open.larksuite.com/open-apis/bot/v2/hook/%s", p.Token)
	if err != nil {
		return fmt.Errorf("build dingtalk url %w", err)
	}
	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("send dingtalk request failed %w", err)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type response struct {
		Code          int         `json:"code"`
		Data          struct{}    `json:"data"`
		Msg           string      `json:"msg"`
		Extra         interface{} `json:"Extra"`
		StatusCode    int         `json:"StatusCode"`
		StatusMessage string      `json:"StatusMessage"`
	}

	var r response
	if err = json.Unmarshal(result, &r); err != nil {
		return fmt.Errorf("parse lark response failed %w", err)
	}

	if r.StatusMessage != "success" {
		return fmt.Errorf("lark message response %w", fmt.Errorf("body: %v", string(result)))
	}
	return nil
}
