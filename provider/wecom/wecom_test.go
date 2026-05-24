package wecom

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

func TestNew_EmptyKeyReturnsNil(t *testing.T) {
	t.Parallel()
	if got := New(Config{}); got != nil {
		t.Errorf("New with empty key returned %v, want nil", got)
	}
}

func TestProvider_Name(t *testing.T) {
	t.Parallel()
	if p := New(Config{Key: "k"}); p.Name() != "wecom" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestSend_TextDefault(t *testing.T) {
	t.Parallel()

	var got textPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "secret-key" {
			t.Errorf("missing key query")
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = io.WriteString(w, `{"errcode":0,"errmsg":"ok"}`)
	}))
	t.Cleanup(srv.Close)

	p := New(Config{Key: "secret-key"})
	p.endpoint = srv.URL

	err := p.Send(context.Background(), provider.Message{
		Subject: "alarm",
		Content: "fire",
		Extras: map[string]any{
			KeyMentionedMobile: []string{"13800000000"},
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got.MsgType != "text" {
		t.Errorf("msgtype = %q", got.MsgType)
	}
	if !strings.Contains(got.Text.Content, "alarm") {
		t.Errorf("content = %q", got.Text.Content)
	}
	if len(got.Text.MentionedMobileList) != 1 || got.Text.MentionedMobileList[0] != "13800000000" {
		t.Errorf("mentioned_mobile_list = %v", got.Text.MentionedMobileList)
	}
}

func TestSend_MarkdownFormat(t *testing.T) {
	t.Parallel()

	var got markdownPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = io.WriteString(w, `{"errcode":0,"errmsg":"ok"}`)
	}))
	t.Cleanup(srv.Close)

	p := New(Config{Key: "k"})
	p.endpoint = srv.URL

	err := p.Send(context.Background(), provider.Message{
		Subject: "alarm",
		Content: "fire",
		Format:  provider.FormatMarkdown,
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got.MsgType != "markdown" {
		t.Errorf("msgtype = %q", got.MsgType)
	}
	if !strings.Contains(got.Markdown.Content, "**alarm**") {
		t.Errorf("markdown content = %q", got.Markdown.Content)
	}
}

func TestSend_BodyError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"errcode":93000,"errmsg":"invalid webhook url"}`)
	}))
	t.Cleanup(srv.Close)

	p := New(Config{Key: "k"})
	p.endpoint = srv.URL

	err := p.Send(context.Background(), provider.Message{Subject: "s"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid webhook url") {
		t.Errorf("err %v missing api message", err)
	}
}

func TestSend_HTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	p := New(Config{Key: "k"})
	p.endpoint = srv.URL

	err := p.Send(context.Background(), provider.Message{Subject: "s"})
	if err == nil || !strings.Contains(err.Error(), "http 500") {
		t.Errorf("err = %v, want http 500", err)
	}
}
