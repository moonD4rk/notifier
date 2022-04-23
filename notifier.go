package notifier

import (
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/moond4rk/notifier/provider/bark"
	"github.com/moond4rk/notifier/provider/dingtalk"
	"github.com/moond4rk/notifier/provider/lark"
)

type Notifier struct {
	Providers []provider
}

func New(options ...Option) *Notifier {
	n := &Notifier{Providers: []provider{}}
	for _, option := range options {
		option(n)
	}
	return n
}

var ErrSendNotification = errors.New("send notification")

func (n *Notifier) Send(subject, content string) error {
	var eg errgroup.Group

	for _, provider := range n.Providers {
		p := provider
		eg.Go(func() error {
			return p.Send(subject, content)
		})
	}
	err := eg.Wait()

	if err != nil {
		err = errors.Wrap(ErrSendNotification, err.Error())
	}

	return err
}

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
