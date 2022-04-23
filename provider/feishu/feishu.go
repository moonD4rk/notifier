package feishu

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
	Token  string
	Secret string
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
	data, err := encrypter.LarkData(subject, content, p.Secret)

	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", p.Token)
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
		return errors.Wrap(err, "parse feishu response failed")
	}

	if r.StatusMessage != "success" {
		err = fmt.Errorf("body: %v", string(result))
		return errors.Wrap(err, "feishu message response error")
	}
	return errors.Wrap(err, "send feishu message failed")
}
