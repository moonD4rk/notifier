// Package wecom implements a notifier.Provider for WeCom (企业微信)
// group robots (https://developer.work.weixin.qq.com/document/path/91770).
package wecom

import (
	"context"
	"fmt"
	"net/http"

	"github.com/moond4rk/notifier/internal/httpx"
	"github.com/moond4rk/notifier/provider"
)

const defaultEndpoint = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send"

// Extras keys consumed by this provider.
//
// MentionedList / MentionedMobileList accept []string. They map to the
// API's mentioned_list and mentioned_mobile_list fields and are only
// honoured when the message is sent as plain text (the default).
const (
	KeyMentionedList   = "wecom.mentioned_list"
	KeyMentionedMobile = "wecom.mentioned_mobile_list"
)

// Config configures a WeCom provider. Key is the webhook key (the value
// of the "key" query parameter on the webhook URL).
type Config struct {
	Key        string
	HTTPClient *http.Client
}

// Provider sends messages to a WeCom group robot.
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
	return &Provider{
		cfg:      cfg,
		endpoint: defaultEndpoint,
		client:   httpx.New(cfg.HTTPClient),
	}
}

// Name implements notifier.Provider.
func (p *Provider) Name() string { return "wecom" }

type markdownPayload struct {
	MsgType  string   `json:"msgtype"`
	Markdown markdown `json:"markdown"`
}

type markdown struct {
	Content string `json:"content"`
}

type textPayload struct {
	MsgType string `json:"msgtype"`
	Text    text   `json:"text"`
}

type text struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

type response struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Send implements notifier.Provider.
func (p *Provider) Send(ctx context.Context, msg provider.Message) error {
	link := fmt.Sprintf("%s?key=%s", p.endpoint, p.cfg.Key)
	body := buildPayload(msg)

	var resp response
	if err := p.client.PostJSON(ctx, link, body, &resp); err != nil {
		return fmt.Errorf("send wecom: %w", err)
	}
	if resp.ErrCode != 0 {
		return fmt.Errorf("wecom: %w (errcode=%d errmsg=%q)",
			httpx.ErrUnexpectedStatus, resp.ErrCode, resp.ErrMsg)
	}
	return nil
}

func buildPayload(msg provider.Message) any {
	if msg.Format == provider.FormatMarkdown {
		return &markdownPayload{
			MsgType:  "markdown",
			Markdown: markdown{Content: fmt.Sprintf("**%s**\n%s", msg.Subject, msg.Content)},
		}
	}
	pd := &textPayload{
		MsgType: "text",
		Text:    text{Content: fmt.Sprintf("%s\n%s", msg.Subject, msg.Content)},
	}
	if v, ok := msg.Extras[KeyMentionedList].([]string); ok {
		pd.Text.MentionedList = v
	}
	if v, ok := msg.Extras[KeyMentionedMobile].([]string); ok {
		pd.Text.MentionedMobileList = v
	}
	return pd
}
