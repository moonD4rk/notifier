package dingtalk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moond4rk/notifier/internal/crypto"
	"github.com/moond4rk/notifier/provider"
)

func TestNew_EmptyTokenReturnsNil(t *testing.T) {
	t.Parallel()
	if got := New(Config{Secret: "secret"}); got != nil {
		t.Errorf("New with empty token returned %v, want nil", got)
	}
}

func TestProvider_Name(t *testing.T) {
	t.Parallel()
	p := New(Config{Token: "tok"})
	if p.Name() != "dingtalk" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestSignedURL_NoSecretOmitsSignature(t *testing.T) {
	t.Parallel()

	p := New(Config{Token: "tok"})
	link, err := p.signedURL()
	if err != nil {
		t.Fatalf("signedURL: %v", err)
	}
	if got := getQuery(t, link, "access_token"); got != "tok" {
		t.Errorf("access_token = %q", got)
	}
	if got := getQuery(t, link, "sign"); got != "" {
		t.Errorf("sign = %q, want empty when no secret", got)
	}
}

func TestSignedURL_WithSecretIncludesSignature(t *testing.T) {
	t.Parallel()

	fixed := time.Unix(1700000000, 0)
	p := New(Config{Token: "tok", Secret: "topsecret"})
	p.now = func() time.Time { return fixed }

	link, err := p.signedURL()
	if err != nil {
		t.Fatalf("signedURL: %v", err)
	}
	if got := getQuery(t, link, "timestamp"); got != strconv.FormatInt(fixed.UnixMilli(), 10) {
		t.Errorf("timestamp = %q", got)
	}
	want := crypto.DingTalkSign(fixed.UnixMilli(), "topsecret")
	if got := getQuery(t, link, "sign"); got != want {
		t.Errorf("sign = %q, want %q", got, want)
	}
}

func TestSend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		msg        provider.Message
		respStatus int
		wantErr    bool
		wantPrefix string
	}{
		{
			name:       "ok plain content",
			msg:        provider.Message{Subject: "subj", Content: "hello"},
			respStatus: http.StatusOK,
			wantPrefix: "### subj",
		},
		{
			name:       "markdown format passes through",
			msg:        provider.Message{Subject: "subj", Content: "**bold**", Format: provider.FormatMarkdown},
			respStatus: http.StatusOK,
			wantPrefix: "**bold**",
		},
		{
			name:       "5xx",
			msg:        provider.Message{Subject: "s"},
			respStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name: "at extras propagate",
			msg: provider.Message{
				Subject: "subj", Content: "body",
				Extras: map[string]any{
					KeyAtAll:     true,
					KeyAtMobiles: []string{"13800000000"},
					KeyAtUserIDs: []string{"u-1"},
				},
			},
			respStatus: http.StatusOK,
			wantPrefix: "### subj",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got payload
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("access_token") != "tok" {
					t.Errorf("missing access_token")
				}
				body, _ := io.ReadAll(r.Body)
				if err := json.Unmarshal(body, &got); err != nil {
					t.Errorf("unmarshal: %v", err)
				}
				w.WriteHeader(tc.respStatus)
			}))
			t.Cleanup(srv.Close)

			p := New(Config{Token: "tok"})
			p.endpoint = srv.URL

			err := p.Send(context.Background(), tc.msg)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got.MsgType != "markdown" {
				t.Errorf("msgtype = %q", got.MsgType)
			}
			if got.Markdown.Title != tc.msg.Subject {
				t.Errorf("title = %q", got.Markdown.Title)
			}
			if !strings.HasPrefix(got.Markdown.Text, tc.wantPrefix) {
				t.Errorf("text = %q, want prefix %q", got.Markdown.Text, tc.wantPrefix)
			}
			if tc.name == "at extras propagate" {
				if got.At == nil || !got.At.IsAtAll {
					t.Errorf("isAtAll not propagated: %+v", got.At)
				}
				if len(got.At.AtMobiles) != 1 || got.At.AtMobiles[0] != "13800000000" {
					t.Errorf("atMobiles = %v", got.At.AtMobiles)
				}
				if len(got.At.AtUserIDs) != 1 || got.At.AtUserIDs[0] != "u-1" {
					t.Errorf("atUserIds = %v", got.At.AtUserIDs)
				}
			}
		})
	}
}

func getQuery(t *testing.T, rawURL, key string) string {
	t.Helper()
	parsed, err := parseQueryFromURL(rawURL)
	if err != nil {
		t.Fatalf("parse %q: %v", rawURL, err)
	}
	return parsed.Get(key)
}
