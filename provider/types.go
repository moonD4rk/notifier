// Package provider defines the Provider interface that every notifier
// channel implements, plus the shared Message and error types.
//
// Concrete provider implementations live in sibling packages
// (provider/bark, provider/dingtalk, ...). The top-level notifier package
// re-exports Provider, Message, and the error types as type aliases so
// callers can write notifier.Provider without an extra import.
package provider

import "context"

// Provider sends a Message to a single notification channel.
type Provider interface {
	// Name identifies the provider for error attribution and metrics.
	// It should be stable, lowercase, and free of whitespace.
	Name() string

	// Send delivers a single message. Implementations must respect ctx
	// for cancellation and timeouts, and should return a non-nil error
	// only when the message was definitely not delivered.
	Send(ctx context.Context, msg Message) error
}

// Format describes the wire format of Message.Content. Providers that do
// not support a particular format treat it as plain text.
type Format string

const (
	FormatText     Format = "text"
	FormatMarkdown Format = "markdown"
)

// Message is a single notification dispatched to one or more providers.
//
// Extras carries provider-specific fields. Each provider package documents
// the keys it consumes (e.g. bark.KeySound, lark.KeyMarkdown); unknown
// keys are silently ignored.
type Message struct {
	Subject string
	Content string
	Format  Format
	Extras  map[string]any
}

// ExtraString returns the string value stored under key in m.Extras, or
// the zero value if the key is absent or holds a non-string type.
func (m Message) ExtraString(key string) string {
	if v, ok := m.Extras[key].(string); ok {
		return v
	}
	return ""
}

// ExtraInt returns the int value stored under key in m.Extras, or the
// zero value if the key is absent or holds a non-int type.
func (m Message) ExtraInt(key string) int {
	if v, ok := m.Extras[key].(int); ok {
		return v
	}
	return 0
}
