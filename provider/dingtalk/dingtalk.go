package dingtalk

import (
	"github.com/CatchZeng/dingtalk/pkg/dingtalk"
	"github.com/pkg/errors"
)

type provider struct {
	Token  string `yaml:"token,omitempty"`
	Secret string `yaml:"secret,omitempty"`
}

func New(token, secret string) *provider {
	if token == "" || secret == "" {
		return nil
	}
	return &provider{
		Token:  token,
		Secret: secret,
	}
}

func (p *provider) Send(subject, content string) error {
	client := dingtalk.NewClient(p.Token, p.Secret)
	msg := dingtalk.NewMarkdownMessage().SetMarkdown(subject, content)
	_, _, err := client.Send(msg)
	return errors.Wrap(err, "send dingtalk message failed")
}
