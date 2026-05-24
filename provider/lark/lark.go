// Package lark implements a notifier.Provider for Lark custom bots
// (https://www.larksuite.com/).
package lark

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/moond4rk/notifier/internal/crypto"
	"github.com/moond4rk/notifier/internal/httpx"
	"github.com/moond4rk/notifier/provider"
)

const defaultEndpoint = "https://open.larksuite.com/open-apis/bot/v2/hook"

// Config configures a Lark provider. Token is the webhook token; Secret
// is only required when the bot has signature verification enabled.
type Config struct {
	Token      string
	Secret     string
	HTTPClient *http.Client
}

// Provider sends messages to a Lark group chat via a custom webhook.
type Provider struct {
	cfg      Config
	endpoint string
	client   *httpx.Client
	now      func() time.Time
}

// New returns a Provider, or nil if cfg.Token is empty.
func New(cfg Config) *Provider {
	if cfg.Token == "" {
		return nil
	}
	return &Provider{
		cfg:      cfg,
		endpoint: defaultEndpoint,
		client:   httpx.New(cfg.HTTPClient),
		now:      time.Now,
	}
}

// Name implements notifier.Provider.
func (p *Provider) Name() string { return "lark" }

type response struct {
	Code          int    `json:"code"`
	Msg           string `json:"msg"`
	StatusCode    int    `json:"StatusCode"`
	StatusMessage string `json:"StatusMessage"`
}

// Send implements notifier.Provider.
func (p *Provider) Send(ctx context.Context, msg provider.Message) error {
	data, err := crypto.LarkPayload(msg.Subject, msg.Content, p.cfg.Secret, p.now())
	if err != nil {
		return fmt.Errorf("build lark payload: %w", err)
	}
	link := fmt.Sprintf("%s/%s", p.endpoint, p.cfg.Token)

	var resp response
	if err := p.client.PostJSON(ctx, link, data, &resp); err != nil {
		return fmt.Errorf("send lark: %w", err)
	}
	if resp.StatusMessage != "success" {
		return fmt.Errorf("lark: %w (status=%q msg=%q code=%d)",
			httpx.ErrUnexpectedStatus, resp.StatusMessage, resp.Msg, resp.Code)
	}
	return nil
}
