package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moond4rk/notifier/provider"
)

func TestNew_RequiresTokenAndChatID(t *testing.T) {
	t.Parallel()

	if got := New(Config{}); got != nil {
		t.Errorf("New with empty fields returned %v, want nil", got)
	}
	if got := New(Config{Token: "t"}); got != nil {
		t.Errorf("New with empty ChatID returned %v, want nil", got)
	}
	if got := New(Config{ChatID: "c"}); got != nil {
		t.Errorf("New with empty Token returned %v, want nil", got)
	}
}

func TestProvider_Name(t *testing.T) {
	t.Parallel()
	p := New(Config{Token: "t", ChatID: "c"})
	if p.Name() != "telegram" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestSend_PlainText(t *testing.T) {
	t.Parallel()

	var got payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want := "/bot123:abc/sendMessage"; r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = io.WriteString(w, `{"ok":true,"result":{}}`)
	}))
	t.Cleanup(srv.Close)

	p := New(Config{Token: "123:abc", ChatID: "@chan"})
	p.endpoint = srv.URL

	err := p.Send(context.Background(), provider.Message{Subject: "alert", Content: "fire"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got.ChatID != "@chan" {
		t.Errorf("chat_id = %q", got.ChatID)
	}
	if got.ParseMode != "" {
		t.Errorf("parse_mode = %q, want empty for text", got.ParseMode)
	}
	if !strings.Contains(got.Text, "alert") || !strings.Contains(got.Text, "fire") {
		t.Errorf("text = %q", got.Text)
	}
}

func TestSend_MarkdownFormat(t *testing.T) {
	t.Parallel()

	var got payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(srv.Close)

	p := New(Config{Token: "t", ChatID: "1"})
	p.endpoint = srv.URL

	err := p.Send(context.Background(), provider.Message{
		Subject: "S", Content: "C",
		Format: provider.FormatMarkdown,
		Extras: map[string]any{
			KeyDisableNotification:   true,
			KeyDisableWebPagePreview: true,
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got.ParseMode != "MarkdownV2" {
		t.Errorf("parse_mode = %q", got.ParseMode)
	}
	if !strings.HasPrefix(got.Text, "*S*") {
		t.Errorf("text = %q", got.Text)
	}
	if !got.DisableNotification {
		t.Error("disable_notification not propagated")
	}
	if !got.DisableWebPagePreview {
		t.Error("disable_web_page_preview not propagated")
	}
}

func TestSend_BodyError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"ok":false,"error_code":400,"description":"chat not found"}`)
	}))
	t.Cleanup(srv.Close)

	p := New(Config{Token: "t", ChatID: "x"})
	p.endpoint = srv.URL

	err := p.Send(context.Background(), provider.Message{Subject: "s"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "chat not found") {
		t.Errorf("err %v missing api description", err)
	}
}

func TestComposeText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  provider.Message
		want string
	}{
		{name: "subject only", msg: provider.Message{Subject: "s"}, want: "s"},
		{name: "content only", msg: provider.Message{Content: "c"}, want: "c"},
		{name: "both text", msg: provider.Message{Subject: "s", Content: "c"}, want: "s\n\nc"},
		{name: "both md", msg: provider.Message{Subject: "s", Content: "c", Format: provider.FormatMarkdown}, want: "*s*\n\nc"},
		{
			name: "md escapes subject and content",
			msg:  provider.Message{Subject: "Build #42 failed.", Content: "Stage 1: ok.", Format: provider.FormatMarkdown},
			want: `*Build \#42 failed\.*` + "\n\n" + `Stage 1: ok\.`,
		},
		{
			name: "md raw passes content through",
			msg: provider.Message{
				Subject: "S.", Content: "*bold* and a.dot",
				Format: provider.FormatMarkdown,
				Extras: map[string]any{KeyRawMarkdown: true},
			},
			want: `*S\.*` + "\n\n*bold* and a.dot",
		},
		{
			name: "md subject only escaped",
			msg:  provider.Message{Subject: "Hi.", Format: provider.FormatMarkdown},
			want: `Hi\.`,
		},
		{
			name: "md content only escaped",
			msg:  provider.Message{Content: "a.b", Format: provider.FormatMarkdown},
			want: `a\.b`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := composeText(tc.msg); got != tc.want {
				t.Errorf("composeText() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEscapeMarkdownV2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{in: "", want: ""},
		{in: "plain", want: "plain"},
		{in: "Build #42 failed.", want: `Build \#42 failed\.`},
		{in: "a_b*c", want: `a\_b\*c`},
		{in: `back\slash`, want: `back\\slash`},
		{in: "中文.test", want: `中文\.test`},
		{in: "all: _*[]()~`>#+-=|{}.!", want: `all: \_\*\[\]\(\)\~\` + "`" + `\>\#\+\-\=\|\{\}\.\!`},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := escapeMarkdownV2(tc.in); got != tc.want {
				t.Errorf("escapeMarkdownV2(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSend_MarkdownEscapesSpecialChars(t *testing.T) {
	t.Parallel()

	var got payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(srv.Close)

	p := New(Config{Token: "t", ChatID: "c"})
	p.endpoint = srv.URL

	err := p.Send(context.Background(), provider.Message{
		Subject: "Alert.1",
		Content: "host=server-01 status=down",
		Format:  provider.FormatMarkdown,
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got.ParseMode != "MarkdownV2" {
		t.Errorf("parse_mode = %q, want MarkdownV2", got.ParseMode)
	}
	wantText := `*Alert\.1*` + "\n\n" + `host\=server\-01 status\=down`
	if got.Text != wantText {
		t.Errorf("text = %q, want %q", got.Text, wantText)
	}
}

func TestSend_RawMarkdownDoesNotEscapeContent(t *testing.T) {
	t.Parallel()

	var got payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(srv.Close)

	p := New(Config{Token: "t", ChatID: "c"})
	p.endpoint = srv.URL

	err := p.Send(context.Background(), provider.Message{
		Subject: "Hello.",
		Content: "*bold* _italic_ `code`",
		Format:  provider.FormatMarkdown,
		Extras:  map[string]any{KeyRawMarkdown: true},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	wantText := `*Hello\.*` + "\n\n" + "*bold* _italic_ `code`"
	if got.Text != wantText {
		t.Errorf("text = %q, want %q", got.Text, wantText)
	}
}
