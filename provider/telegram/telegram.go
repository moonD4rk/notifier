// Package telegram implements a notifier.Provider for Telegram Bot API
// (https://core.telegram.org/bots/api#sendmessage).
package telegram

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/moond4rk/notifier/internal/httpx"
	"github.com/moond4rk/notifier/provider"
)

const defaultEndpoint = "https://api.telegram.org"

// Extras keys consumed by this provider.
const (
	// KeyDisableNotification accepts bool — silent message, recipients
	// receive a notification with no sound.
	KeyDisableNotification = "telegram.disable_notification"
	// KeyDisableWebPagePreview accepts bool — disable link previews.
	KeyDisableWebPagePreview = "telegram.disable_web_page_preview"
	// KeyRawMarkdown accepts bool. When true and Format is Markdown, the
	// library does not escape MarkdownV2 special characters in
	// Message.Content; the caller is then responsible for providing
	// well-formed MarkdownV2. Subject is always escaped because the
	// library wraps it in '*...*' formatting.
	KeyRawMarkdown = "telegram.raw_markdown"
)

// markdownV2Specials lists every character that must be backslash-escaped
// inside a Telegram MarkdownV2 message body. See
// https://core.telegram.org/bots/api#markdownv2-style.
const markdownV2Specials = "_*[]()~`>#+-=|{}.!\\"

// escapeMarkdownV2 backslash-escapes any character in s that has special
// meaning in Telegram's MarkdownV2 grammar.
func escapeMarkdownV2(s string) string {
	if !strings.ContainsAny(s, markdownV2Specials) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 8)
	for _, r := range s {
		if strings.ContainsRune(markdownV2Specials, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Config configures a Telegram provider. Token is the bot token (from
// @BotFather); ChatID is the destination chat (numeric ID, channel
// @username, or "@channelname"). Both are required.
type Config struct {
	Token      string
	ChatID     string
	HTTPClient *http.Client
}

// Provider sends messages to a Telegram chat via the Bot API.
type Provider struct {
	cfg      Config
	endpoint string
	client   *httpx.Client
}

// New returns a Provider, or nil if Token or ChatID is empty.
func New(cfg Config) *Provider {
	if cfg.Token == "" || cfg.ChatID == "" {
		return nil
	}
	return &Provider{
		cfg:      cfg,
		endpoint: defaultEndpoint,
		client:   httpx.New(cfg.HTTPClient),
	}
}

// Name implements notifier.Provider.
func (p *Provider) Name() string { return "telegram" }

type payload struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableNotification   bool   `json:"disable_notification,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
}

type response struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

// Send implements notifier.Provider.
func (p *Provider) Send(ctx context.Context, msg provider.Message) error {
	link := fmt.Sprintf("%s/bot%s/sendMessage", p.endpoint, p.cfg.Token)
	pd := buildPayload(p.cfg.ChatID, msg)

	var resp response
	if err := p.client.PostJSON(ctx, link, pd, &resp); err != nil {
		return fmt.Errorf("send telegram: %w", err)
	}
	if !resp.OK {
		return fmt.Errorf("telegram: %w (error_code=%d description=%q)",
			httpx.ErrUnexpectedStatus, resp.ErrorCode, resp.Description)
	}
	return nil
}

func buildPayload(chatID string, msg provider.Message) *payload {
	pd := &payload{ChatID: chatID, Text: composeText(msg)}
	if msg.Format == provider.FormatMarkdown {
		pd.ParseMode = "MarkdownV2"
	}
	if v, ok := msg.Extras[KeyDisableNotification].(bool); ok {
		pd.DisableNotification = v
	}
	if v, ok := msg.Extras[KeyDisableWebPagePreview].(bool); ok {
		pd.DisableWebPagePreview = v
	}
	return pd
}

func composeText(msg provider.Message) string {
	if msg.Format != provider.FormatMarkdown {
		switch {
		case msg.Subject != "" && msg.Content != "":
			return fmt.Sprintf("%s\n\n%s", msg.Subject, msg.Content)
		case msg.Subject != "":
			return msg.Subject
		default:
			return msg.Content
		}
	}

	subject := escapeMarkdownV2(msg.Subject)
	content := msg.Content
	if raw, _ := msg.Extras[KeyRawMarkdown].(bool); !raw {
		content = escapeMarkdownV2(content)
	}
	switch {
	case subject != "" && content != "":
		return fmt.Sprintf("*%s*\n\n%s", subject, content)
	case subject != "":
		return subject
	default:
		return content
	}
}
