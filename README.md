# notifier

[![Build](https://github.com/moonD4rk/notifier/actions/workflows/build.yml/badge.svg)](https://github.com/moonD4rk/notifier/actions/workflows/build.yml)
[![Tests](https://github.com/moonD4rk/notifier/actions/workflows/test.yml/badge.svg)](https://github.com/moonD4rk/notifier/actions/workflows/test.yml)
[![Lint](https://github.com/moonD4rk/notifier/actions/workflows/lint.yml/badge.svg)](https://github.com/moonD4rk/notifier/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/moond4rk/notifier.svg)](https://pkg.go.dev/github.com/moond4rk/notifier)
[![Coverage Status](https://coveralls.io/repos/github/moonD4rk/notifier/badge.svg?branch=main)](https://coveralls.io/github/moonD4rk/notifier?branch=main)

A small, zero-dependency Go library that fans a single message out to multiple notification channels concurrently. Useful for alerting, build pipelines, personal automations, and anywhere you want one call to reach DingTalk, Bark, Telegram, etc. without writing five HTTP clients.

## Providers

| Provider | Package |
|---|---|
| [DingTalk](https://www.dingtalk.com/) | [`provider/dingtalk`](./provider/dingtalk) |
| [Bark](https://apps.apple.com/us/app/bark-customed-notifications/id1403753865) | [`provider/bark`](./provider/bark) |
| [Lark](https://www.larksuite.com/) | [`provider/lark`](./provider/lark) |
| [Feishu](https://www.feishu.cn/) | [`provider/feishu`](./provider/feishu) |
| [Server 酱](https://sct.ftqq.com/) | [`provider/serverchan`](./provider/serverchan) |
| [WeCom (企业微信)](https://work.weixin.qq.com/) | [`provider/wecom`](./provider/wecom) |
| [Telegram Bot](https://core.telegram.org/bots) | [`provider/telegram`](./provider/telegram) |

## Install

```sh
go get github.com/moond4rk/notifier
```

Requires Go 1.26 or later.

## Quickstart

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/moond4rk/notifier"
)

func main() {
    n := notifier.New(
        notifier.WithDingTalk(os.Getenv("DINGTALK_TOKEN"), os.Getenv("DINGTALK_SECRET")),
        notifier.WithBark(os.Getenv("BARK_KEY"), ""),
        notifier.WithTelegram(os.Getenv("TELEGRAM_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID")),
    )

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := n.Send(ctx, notifier.Message{
        Subject: "build failed",
        Content: "main.go:42 syntax error",
    }); err != nil {
        log.Print(err)
    }
}
```

A failure in one provider does not silence the rest — `Send` returns a `*notifier.MultiError` that wraps each per-provider failure as a `*notifier.SendError`. `errors.As` and `errors.Is` work as you'd expect:

```go
var multi *notifier.MultiError
if errors.As(err, &multi) {
    for _, e := range multi.Errors {
        log.Printf("%s: %v", e.Provider, e.Err)
    }
}
```

## Channel-specific options (Extras)

`Message.Extras` is a `map[string]any` of provider-specific keys. Unknown keys are silently ignored, so it is safe to set them when broadcasting to a mixed list of providers.

```go
import (
    "github.com/moond4rk/notifier"
    "github.com/moond4rk/notifier/provider/bark"
    "github.com/moond4rk/notifier/provider/dingtalk"
    "github.com/moond4rk/notifier/provider/telegram"
)

n.Send(ctx, notifier.Message{
    Subject: "alarm",
    Content: "fire in srv-01",
    Format:  notifier.FormatMarkdown,
    Extras: map[string]any{
        bark.KeySound:               "siren.caf",
        bark.KeyGroup:               "alerts",
        dingtalk.KeyAtAll:           true,
        telegram.KeyDisableNotification: false,
    },
})
```

| Provider | Key constant | Type | Effect |
|---|---|---|---|
| `bark` | `KeySound` | `string` | Override notification sound |
| `bark` | `KeyIcon` | `string` | Custom icon URL |
| `bark` | `KeyGroup` | `string` | Group key |
| `bark` | `KeyURL` | `string` | Tap-to-open URL |
| `bark` | `KeyBadge` | `int` | Badge count |
| `dingtalk` | `KeyAtAll` | `bool` | @-everyone |
| `dingtalk` | `KeyAtMobiles` | `[]string` | @ by phone number |
| `dingtalk` | `KeyAtUserIDs` | `[]string` | @ by user ID |
| `wecom` | `KeyMentionedList` | `[]string` | @ by username (text mode) |
| `wecom` | `KeyMentionedMobile` | `[]string` | @ by phone (text mode) |
| `telegram` | `KeyDisableNotification` | `bool` | Silent message |
| `telegram` | `KeyDisableWebPagePreview` | `bool` | No link preview |

## Custom providers

Implement the `notifier.Provider` interface (alias of `provider.Provider`) and register with `WithProvider`:

```go
type slackProvider struct{ webhook string }

func (slackProvider) Name() string { return "slack" }

func (s slackProvider) Send(ctx context.Context, msg notifier.Message) error {
    // POST to s.webhook with ctx; return error on failure.
    return nil
}

n := notifier.New(notifier.WithProvider(slackProvider{webhook: "..."}))
```

## HTTP client and timeouts

Each built-in provider uses an `*http.Client` with a 30-second timeout by default. To share a client (for connection pooling, proxies, custom transports, ...), pass it through `WithHTTPClient` *before* the convenience helpers:

```go
client := &http.Client{Timeout: 5 * time.Second}
n := notifier.New(
    notifier.WithHTTPClient(client),         // must come first
    notifier.WithDingTalk(token, secret),
    notifier.WithBark(key, ""),
)
```

Or set it per provider via the `Config` struct when constructing manually:

```go
notifier.WithProvider(dingtalk.New(dingtalk.Config{
    Token: token, Secret: secret, HTTPClient: client,
}))
```

## License

MIT — see [LICENSE](./LICENSE).
