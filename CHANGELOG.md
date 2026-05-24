# Changelog

All notable changes to this project are documented here. The format is
based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Breaking

- `Notifier.Send` now takes `context.Context` and a `notifier.Message`:
  `Send(ctx context.Context, msg Message) error`. The previous
  `Send(subject, content string) error` signature has been removed.
- The unexported `provider` interface is replaced by the exported
  `notifier.Provider` (alias of `notifier/provider.Provider`) which
  requires a `Name()` method. Custom providers must add `Name()`.
- Each `provider/*` package's `New(...)` constructor now takes a
  `Config` struct instead of positional arguments. Existing callers
  using `notifier.WithDingTalk(token, secret)` etc. continue to work
  unchanged.
- `notifier.DefaultBarkServer` is removed; use `bark.DefaultServer`.
- The previous error sentinel `ErrSendNotification` is replaced by
  `*notifier.MultiError`, which wraps per-provider `*SendError` values
  (`errors.As`/`errors.Is` work as expected).
- Module now requires Go 1.22 and has no third-party dependencies
  (`golang.org/x/sync` removed; broadcast uses `sync.WaitGroup`).
- The `_example/` directory has been replaced by `examples/`.

### Added

- `notifier.WithProvider(Provider)` for registering arbitrary providers.
- `notifier.WithHTTPClient(*http.Client)` to share a client across the
  built-in convenience helpers.
- New providers: **WeCom** (`provider/wecom`, registered via
  `notifier.WithWeCom(key)`) and **Telegram Bot** (`provider/telegram`,
  registered via `notifier.WithTelegram(token, chatID)`).
- `notifier.Message.Extras` for channel-specific fields:
  - **Bark**: `KeySound`, `KeyIcon`, `KeyGroup`, `KeyURL`, `KeyBadge`
  - **DingTalk**: `KeyAtMobiles`, `KeyAtUserIDs`, `KeyAtAll` for @-mentions
  - **WeCom**: `KeyMentionedList`, `KeyMentionedMobile`
  - **Telegram**: `KeyDisableNotification`, `KeyDisableWebPagePreview`,
    `KeyRawMarkdown` (opt out of MarkdownV2 escaping for Content)
- `notifier.MultiError`, `notifier.SendError` for structured failure
  reporting.
- Per-provider `Config` struct with optional `HTTPClient`.
- `internal/httpx` shared POST/JSON helper with 30s default timeout and
  1 MiB response cap.
- Pure functions `crypto.DingTalkSign` and `crypto.LarkSign` /
  `crypto.LarkPayload(now)` for deterministic, testable signatures.
- httptest-based unit tests for every provider (~90% coverage),
  including signature, request body, and error handling.
- golangci-lint v2 configuration; CI runs `go test -race -covermode=atomic`.

### Changed

- HTTP requests now go through `http.NewRequestWithContext` with a
  default 30-second timeout. Server response bodies are read through
  `io.LimitReader` to cap at 1 MiB.
- Lark/Feishu error messages now surface the body's `StatusMessage`
  field; previously only the often-empty `msg` field was shown.
- Telegram MarkdownV2 messages now have their Subject and Content
  automatically backslash-escaped, so values containing characters
  like `.`, `_`, or `-` no longer trigger Telegram parse errors. Pass
  `Extras[telegram.KeyRawMarkdown] = true` to send pre-escaped
  MarkdownV2 content.

### Security

- Webhook tokens embedded in request URLs (DingTalk, Feishu, Lark,
  ServerChan, Telegram, WeCom) are now stripped from `*url.Error`
  values returned by `internal/httpx` so they cannot leak into logs
  via error wrapping. `errors.Is` against context sentinels continues
  to work.

### Removed

- `notifier.ErrSendNotification` (replaced by `*MultiError`).
- `notifier.DefaultBarkServer` (moved to `bark.DefaultServer`).
- Top-level `consts.go` and `provider.go` (moved into the appropriate
  subpackages).
- `golang.org/x/sync` dependency.
