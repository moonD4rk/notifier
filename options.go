package notifier

import (
	"github.com/moond4rk/notifier/provider/bark"
	"github.com/moond4rk/notifier/provider/dingtalk"
	"github.com/moond4rk/notifier/provider/feishu"
	"github.com/moond4rk/notifier/provider/lark"
	"github.com/moond4rk/notifier/provider/serverchan"
)

type Option func(p *Notifier)

func WithDingTalk(token, secret string) Option {
	d := dingtalk.New(token, secret)
	return func(n *Notifier) {
		if d != nil {
			n.Providers = append(n.Providers, d)
		}
	}
}

func WithBark(key, server string) Option {
	b := bark.New(key, server)
	return func(n *Notifier) {
		if b != nil {
			n.Providers = append(n.Providers, b)
		}
	}
}

func WithLark(token, secret string) Option {
	l := lark.New(token, secret)
	return func(n *Notifier) {
		if l != nil {
			n.Providers = append(n.Providers, l)
		}
	}
}

func WithFeishu(token, secret string) Option {
	l := feishu.New(token, secret)
	return func(n *Notifier) {
		if l != nil {
			n.Providers = append(n.Providers, l)
		}
	}
}

func WithServerChan(userID, sendKey string) Option {
	s := serverchan.New(userID, sendKey)
	return func(n *Notifier) {
		if s != nil {
			n.Providers = append(n.Providers, s)
		}
	}
}
