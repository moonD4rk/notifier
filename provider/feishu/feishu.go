// Package feishu implements a notifier.Provider for Feishu custom bots
// (https://www.feishu.cn/).
package feishu

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/moond4rk/notifier/internal/crypto"
	"github.com/moond4rk/notifier/internal/httpx"
	"github.com/moond4rk/notifier/provider"
)

const defaultEndpoint = "https://open.feishu.cn/open-apis/bot/v2/hook"

// Config configures a Feishu provider. Token is the webhook token;
// Secret is only required when signature verification is enabled.
type Config struct {
	Token      string
	Secret     string
	HTTPClient *http.Client
}

// Provider sends messages to a Feishu group chat via a custom webhook.
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
func (p *Provider) Name() string { return "feishu" }

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
		return fmt.Errorf("build feishu payload: %w", err)
	}
	link := fmt.Sprintf("%s/%s", p.endpoint, p.cfg.Token)

	var resp response
	if err := p.client.PostJSON(ctx, link, data, &resp); err != nil {
		return fmt.Errorf("send feishu: %w", err)
	}
	if resp.StatusMessage != "success" {
		return fmt.Errorf("feishu: %w (status=%q msg=%q code=%d)",
			httpx.ErrUnexpectedStatus, resp.StatusMessage, resp.Msg, resp.Code)
	}
	return nil
}
