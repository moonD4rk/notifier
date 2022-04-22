package notifier

import (
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/moond4rk/notifier/provider/bark"
	"github.com/moond4rk/notifier/provider/dingtalk"
)

type Notifier struct {
	Providers []Provider
}

type Provider interface {
	Send(title, content string) error
}

func New(options ...Option) *Notifier {
	n := &Notifier{Providers: []Provider{}}
	for _, option := range options {
		option(n)
	}
	return n
}

var ErrSendNotification = errors.New("send notification")

func (n *Notifier) Send(subject, content string) error {
	var eg errgroup.Group

	for _, service := range n.Providers {
		if service != nil {
			s := service
			eg.Go(func() error {
				return s.Send(subject, content)
			})
		}
	}
	err := eg.Wait()

	if err != nil {
		err = errors.Wrap(ErrSendNotification, err.Error())
	}

	return err
}

type Option func(p *Notifier)

func WithDingTalk(token, secret string) Option {
	ding := dingtalk.New(token, secret)
	return func(p *Notifier) {
		p.Providers = append(p.Providers, ding)
	}
}

func WithBark(key, server string) Option {
	bark := bark.New(key, server)
	return func(p *Notifier) {
		p.Providers = append(p.Providers, bark)
	}
}
