// Package serverchan implements a notifier.Provider for Server 酱
// (https://sct.ftqq.com/).
package serverchan

import (
	"context"
	"fmt"
	"net/http"

	"github.com/moond4rk/notifier/internal/httpx"
	"github.com/moond4rk/notifier/provider"
)

const defaultEndpoint = "https://sctapi.ftqq.com"

// Config configures a Server 酱 provider. SendKey is required; UserID
// can be empty for the modern SCT (Server酱·Turbo) API.
type Config struct {
	UserID     string
	SendKey    string
	HTTPClient *http.Client
}

// Provider sends messages to a Server 酱 endpoint.
type Provider struct {
	cfg      Config
	endpoint string
	client   *httpx.Client
}

// New returns a Provider, or nil if cfg.SendKey is empty.
func New(cfg Config) *Provider {
	if cfg.SendKey == "" {
		return nil
	}
	return &Provider{
		cfg:      cfg,
		endpoint: defaultEndpoint,
		client:   httpx.New(cfg.HTTPClient),
	}
}

// Name implements notifier.Provider.
func (p *Provider) Name() string { return "serverchan" }

type payload struct {
	Text string `json:"text"`
	Desp string `json:"desp"`
}

// Send implements notifier.Provider.
func (p *Provider) Send(ctx context.Context, msg provider.Message) error {
	link := fmt.Sprintf("%s/%s.send", p.endpoint, p.cfg.SendKey)
	pd := &payload{Text: msg.Subject, Desp: msg.Content}
	if err := p.client.PostJSON(ctx, link, pd, nil); err != nil {
		return fmt.Errorf("send serverchan: %w", err)
	}
	return nil
}
