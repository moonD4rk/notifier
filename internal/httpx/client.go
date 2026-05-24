// Package httpx provides a small HTTP helper shared by all notifier providers.
//
// It centralises request construction, body handling, response decoding, and
// error wrapping so individual providers can focus on their wire format.
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	maxBodyBytes   = 1 << 20
)

// Client wraps an *http.Client with the convenience method PostJSON.
//
// The zero value is not safe to use; call New.
type Client struct {
	HTTP *http.Client
}

// New returns a Client that wraps the provided *http.Client. If hc is nil, a
// default client with a 30-second timeout is used.
func New(hc *http.Client) *Client {
	if hc == nil {
		hc = &http.Client{Timeout: defaultTimeout}
	}
	return &Client{HTTP: hc}
}

// PostJSON sends a POST request with a JSON body to requestURL. The body
// argument may be a []byte or io.Reader (used as-is) or any other value
// (marshalled with encoding/json). If out is non-nil and the response is
// successful, the response body is unmarshalled into out.
//
// The response body is capped at 1 MiB to avoid runaway reads. A non-2xx
// status returns an error containing the status code and (truncated) body.
//
// Network errors are redacted so that path and query components of the
// request URL never appear in returned error strings; this prevents
// webhook tokens embedded in URLs from leaking via logs.
func (c *Client) PostJSON(ctx context.Context, requestURL string, body, out any) error {
	reader, err := bodyReader(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, reader)
	if err != nil {
		return fmt.Errorf("new request: %w", redactURLError(err))
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", redactURLError(err))
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("http %d: %s", resp.StatusCode, raw)
	}

	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}

func bodyReader(body any) (io.Reader, error) {
	switch b := body.(type) {
	case nil:
		return nil, nil
	case []byte:
		return bytes.NewReader(b), nil
	case io.Reader:
		return b, nil
	default:
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		return bytes.NewReader(buf), nil
	}
}

// ErrUnexpectedStatus is returned when a provider's response payload signals
// failure even though the HTTP status was 2xx (e.g. Lark/Feishu return
// status_message != "success").
var ErrUnexpectedStatus = errors.New("unexpected response status")

// redactURLError returns a copy of err with the URL component of any wrapped
// *url.Error reduced to scheme://host. The underlying Err is preserved so
// errors.Is against sentinels like context.Canceled still matches.
func redactURLError(err error) error {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return err
	}
	redacted := "[redacted]"
	if u, parseErr := url.Parse(urlErr.URL); parseErr == nil && u.Host != "" {
		redacted = u.Scheme + "://" + u.Host
	}
	return &url.Error{Op: urlErr.Op, URL: redacted, Err: urlErr.Err}
}
