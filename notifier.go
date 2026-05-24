package notifier

import (
	"context"
	"net/http"
	"sync"

	"github.com/moond4rk/notifier/provider"
)

// Provider is an alias of provider.Provider, the interface every channel
// implements. Anything implementing it can be registered with WithProvider.
type Provider = provider.Provider

// Message is an alias of provider.Message — the payload broadcast to
// providers.
type Message = provider.Message

// Format is an alias of provider.Format, describing the wire format of
// Message.Content.
type Format = provider.Format

// SendError is an alias of provider.SendError, returned per-provider in a
// MultiError.
type SendError = provider.SendError

// MultiError is an alias of provider.MultiError aggregating per-provider
// failures from a single Send call.
type MultiError = provider.MultiError

// FormatText is plain text, the default Format.
const FormatText = provider.FormatText

// FormatMarkdown signals the message body uses Markdown.
const FormatMarkdown = provider.FormatMarkdown

// ErrNoProvider is returned by Send when no providers are configured.
var ErrNoProvider = provider.ErrNoProvider

// Notifier dispatches a Message to all configured providers concurrently.
// Use New to construct one and Send to deliver a message.
type Notifier struct {
	providers  []Provider
	httpClient *http.Client
}

// New returns a Notifier configured by the supplied options.
func New(options ...Option) *Notifier {
	n := &Notifier{}
	for _, opt := range options {
		opt(n)
	}
	return n
}

// HTTPClient returns the shared *http.Client used by convenience helpers
// such as WithDingTalk to construct providers. Returns nil if
// WithHTTPClient was never called; provider packages then fall back to
// their own default client.
func (n *Notifier) HTTPClient() *http.Client { return n.httpClient }

// Providers returns a copy of the configured providers slice.
func (n *Notifier) Providers() []Provider {
	out := make([]Provider, len(n.providers))
	copy(out, n.providers)
	return out
}

// Send dispatches msg to every configured provider concurrently. It returns:
//
//   - ErrNoProvider if no providers are configured
//   - *MultiError if at least one provider failed; successful providers
//     are not reported
//   - nil if all providers succeeded
//
// ctx is forwarded unchanged to each provider; cancelling it aborts every
// in-flight request. A failure in one provider does not prevent the
// others from running.
func (n *Notifier) Send(ctx context.Context, msg Message) error {
	if len(n.providers) == 0 {
		return ErrNoProvider
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []*SendError
	)
	for _, p := range n.providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			if err := p.Send(ctx, msg); err != nil {
				mu.Lock()
				errs = append(errs, &SendError{Provider: p.Name(), Err: err})
				mu.Unlock()
			}
		}(p)
	}
	wg.Wait()

	if len(errs) > 0 {
		return &MultiError{Errors: errs}
	}
	return nil
}
