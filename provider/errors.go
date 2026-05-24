package provider

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNoProvider is returned by Notifier.Send when no providers are
// configured.
var ErrNoProvider = errors.New("notifier: no provider configured")

// SendError reports a single provider's failure. The Provider field holds
// the value returned by Provider.Name().
type SendError struct {
	Provider string
	Err      error
}

func (e *SendError) Error() string { return fmt.Sprintf("%s: %v", e.Provider, e.Err) }

func (e *SendError) Unwrap() error { return e.Err }

// MultiError aggregates per-provider failures from a single Send call.
// Use errors.As to extract individual *SendError values, or errors.Is to
// test against a sentinel like context.DeadlineExceeded.
type MultiError struct {
	Errors []*SendError
}

func (m *MultiError) Error() string {
	if m == nil || len(m.Errors) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m.Errors))
	for _, e := range m.Errors {
		parts = append(parts, e.Error())
	}
	return fmt.Sprintf("%d provider(s) failed: %s", len(m.Errors), strings.Join(parts, "; "))
}

// Unwrap returns the underlying SendErrors. Combined with Go 1.20+
// multi-error semantics, this lets callers use errors.Is/As to drill
// into individual failures.
func (m *MultiError) Unwrap() []error {
	out := make([]error, len(m.Errors))
	for i, e := range m.Errors {
		out[i] = e
	}
	return out
}
