// Package dingtalk implements a notifier.Provider for DingTalk group
// robots (https://open.dingtalk.com/).
package dingtalk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/moond4rk/notifier/internal/crypto"
	"github.com/moond4rk/notifier/internal/httpx"
	"github.com/moond4rk/notifier/provider"
)

const defaultEndpoint = "https://oapi.dingtalk.com/robot/send"

// Extras keys consumed by this provider. AtMobiles / AtUserIDs accept
// []string; AtAll accepts bool.
const (
	KeyAtMobiles = "dingtalk.at_mobiles"
	KeyAtUserIDs = "dingtalk.at_user_ids"
	KeyAtAll     = "dingtalk.at_all"
)

// Config configures a DingTalk provider. Token is required; Secret is
// only required for bots configured with signed-request security.
type Config struct {
	Token      string
	Secret     string
	HTTPClient *http.Client
}

// Provider sends messages to a DingTalk group robot.
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
func (p *Provider) Name() string { return "dingtalk" }

// Send implements notifier.Provider.
func (p *Provider) Send(ctx context.Context, msg provider.Message) error {
	link, err := p.signedURL()
	if err != nil {
		return fmt.Errorf("build dingtalk url: %w", err)
	}
	if err := p.client.PostJSON(ctx, link, buildPayload(msg), nil); err != nil {
		return fmt.Errorf("send dingtalk: %w", err)
	}
	return nil
}

func (p *Provider) signedURL() (string, error) {
	u, err := url.Parse(p.endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("access_token", p.cfg.Token)
	if p.cfg.Secret != "" {
		ts := p.now().UnixMilli()
		q.Set("timestamp", strconv.FormatInt(ts, 10))
		q.Set("sign", crypto.DingTalkSign(ts, p.cfg.Secret))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type payload struct {
	MsgType  string   `json:"msgtype"`
	Markdown markdown `json:"markdown"`
	At       *atSpec  `json:"at,omitempty"`
}

type markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type atSpec struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	AtUserIDs []string `json:"atUserIds,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

func buildPayload(msg provider.Message) *payload {
	text := msg.Content
	if msg.Format != provider.FormatMarkdown {
		text = fmt.Sprintf("### %s\n>%s", msg.Subject, msg.Content)
	}
	pd := &payload{
		MsgType:  "markdown",
		Markdown: markdown{Title: msg.Subject, Text: text},
	}
	if at := buildAt(msg); at != nil {
		pd.At = at
	}
	return pd
}

func buildAt(msg provider.Message) *atSpec {
	at := &atSpec{}
	if v, ok := msg.Extras[KeyAtMobiles].([]string); ok {
		at.AtMobiles = v
	}
	if v, ok := msg.Extras[KeyAtUserIDs].([]string); ok {
		at.AtUserIDs = v
	}
	if v, ok := msg.Extras[KeyAtAll].(bool); ok {
		at.IsAtAll = v
	}
	if len(at.AtMobiles) == 0 && len(at.AtUserIDs) == 0 && !at.IsAtAll {
		return nil
	}
	return at
}
