package notifier

import (
	"errors"

	"golang.org/x/sync/errgroup"
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
var ErrNoProvider = errors.New("no provider, please check your config")

func (n *Notifier) Send(subject, content string) error {
	if len(n.Providers) == 0 {
		return ErrNoProvider
	}
	var eg errgroup.Group
	for _, provider := range n.Providers {
		p := provider
		eg.Go(func() error {
			return p.Send(subject, content)
		})
	}
	err := eg.Wait()

	if err != nil {
		err = errors.Join(ErrSendNotification, err)
	}

	return err
}
