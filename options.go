package notifier

import (
	"net/http"

	"github.com/moond4rk/notifier/provider/bark"
	"github.com/moond4rk/notifier/provider/dingtalk"
	"github.com/moond4rk/notifier/provider/feishu"
	"github.com/moond4rk/notifier/provider/lark"
	"github.com/moond4rk/notifier/provider/serverchan"
	"github.com/moond4rk/notifier/provider/telegram"
	"github.com/moond4rk/notifier/provider/wecom"
)

// Option configures a Notifier. Options are applied in order, so
// WithHTTPClient must precede any WithXxx convenience helper that should
// use the shared client.
type Option func(*Notifier)

// WithProvider registers a Provider. A nil Provider is silently ignored
// so callers can pass the result of pkg.New directly even when its
// required field is empty.
func WithProvider(p Provider) Option {
	return func(n *Notifier) {
		if p != nil {
			n.providers = append(n.providers, p)
		}
	}
}

// WithHTTPClient sets the *http.Client used by the convenience helpers
// (WithDingTalk, WithBark, ...) when they construct providers. Providers
// added via WithProvider keep whatever client they were configured with.
func WithHTTPClient(hc *http.Client) Option {
	return func(n *Notifier) { n.httpClient = hc }
}

// WithDingTalk registers a DingTalk webhook provider with the given
// access token and (optional) secret.
func WithDingTalk(token, secret string) Option {
	return func(n *Notifier) {
		if p := dingtalk.New(dingtalk.Config{Token: token, Secret: secret, HTTPClient: n.httpClient}); p != nil {
			n.providers = append(n.providers, p)
		}
	}
}

// WithBark registers a Bark provider. server is the host (e.g.
// "api.day.app"); pass an empty string to use the public default.
func WithBark(key, server string) Option {
	return func(n *Notifier) {
		if p := bark.New(bark.Config{Key: key, Server: server, HTTPClient: n.httpClient}); p != nil {
			n.providers = append(n.providers, p)
		}
	}
}

// WithLark registers a Lark webhook provider.
func WithLark(token, secret string) Option {
	return func(n *Notifier) {
		if p := lark.New(lark.Config{Token: token, Secret: secret, HTTPClient: n.httpClient}); p != nil {
			n.providers = append(n.providers, p)
		}
	}
}

// WithFeishu registers a Feishu webhook provider.
func WithFeishu(token, secret string) Option {
	return func(n *Notifier) {
		if p := feishu.New(feishu.Config{Token: token, Secret: secret, HTTPClient: n.httpClient}); p != nil {
			n.providers = append(n.providers, p)
		}
	}
}

// WithServerChan registers a Server 酱 provider. userID may be empty for
// the modern SCT API.
func WithServerChan(userID, sendKey string) Option {
	return func(n *Notifier) {
		if p := serverchan.New(serverchan.Config{UserID: userID, SendKey: sendKey, HTTPClient: n.httpClient}); p != nil {
			n.providers = append(n.providers, p)
		}
	}
}

// WithWeCom registers a WeCom (企业微信) group robot provider. key is
// the value of the "key" query parameter on the bot's webhook URL.
func WithWeCom(key string) Option {
	return func(n *Notifier) {
		if p := wecom.New(wecom.Config{Key: key, HTTPClient: n.httpClient}); p != nil {
			n.providers = append(n.providers, p)
		}
	}
}

// WithTelegram registers a Telegram bot provider. token is the bot
// token from @BotFather; chatID is the destination chat (numeric or
// "@channel" form).
func WithTelegram(token, chatID string) Option {
	return func(n *Notifier) {
		if p := telegram.New(telegram.Config{Token: token, ChatID: chatID, HTTPClient: n.httpClient}); p != nil {
			n.providers = append(n.providers, p)
		}
	}
}
