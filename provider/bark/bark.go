// Package bark implements a notifier.Provider for the Bark iOS app
// (https://day.app/).
package bark

import (
	"context"
	"fmt"
	"net/http"

	"github.com/moond4rk/notifier/internal/httpx"
	"github.com/moond4rk/notifier/provider"
)

// DefaultServer is the host of the public Bark server.
const DefaultServer = "api.day.app"

// Extras keys consumed by this provider.
const (
	KeySound = "bark.sound"
	KeyIcon  = "bark.icon"
	KeyGroup = "bark.group"
	KeyURL   = "bark.url"
	KeyBadge = "bark.badge"
)

const defaultSound = "alarm.caf"

// Config configures a Bark provider. Key (the device key) is required;
// Server and HTTPClient are optional.
type Config struct {
	Key        string
	Server     string
	HTTPClient *http.Client
}

// Provider sends messages to a Bark device via the /push endpoint.
type Provider struct {
	cfg      Config
	endpoint string
	client   *httpx.Client
}

// New returns a Provider, or nil if cfg.Key is empty.
func New(cfg Config) *Provider {
	if cfg.Key == "" {
		return nil
	}
	host := cfg.Server
	if host == "" {
		host = DefaultServer
	}
	return &Provider{
		cfg:      cfg,
		endpoint: "https://" + host,
		client:   httpx.New(cfg.HTTPClient),
	}
}

// Name implements notifier.Provider.
func (p *Provider) Name() string { return "bark" }

type payload struct {
	DeviceKey string `json:"device_key"`
	Title     string `json:"title"`
	Body      string `json:"body,omitempty"`
	Badge     int    `json:"badge,omitempty"`
	Sound     string `json:"sound,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Group     string `json:"group,omitempty"`
	URL       string `json:"link,omitempty"`
}

// Send implements notifier.Provider.
func (p *Provider) Send(ctx context.Context, msg provider.Message) error {
	pd := &payload{
		DeviceKey: p.cfg.Key,
		Title:     msg.Subject,
		Body:      msg.Content,
		Sound:     defaultSound,
	}
	if v := msg.ExtraString(KeySound); v != "" {
		pd.Sound = v
	}
	if v := msg.ExtraString(KeyIcon); v != "" {
		pd.Icon = v
	}
	if v := msg.ExtraString(KeyGroup); v != "" {
		pd.Group = v
	}
	if v := msg.ExtraString(KeyURL); v != "" {
		pd.URL = v
	}
	if v := msg.ExtraInt(KeyBadge); v != 0 {
		pd.Badge = v
	}

	if err := p.client.PostJSON(ctx, p.endpoint+"/push", pd, nil); err != nil {
		return fmt.Errorf("send bark: %w", err)
	}
	return nil
}
