package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPostJSON_SuccessSendsBody(t *testing.T) {
	t.Parallel()

	type request struct {
		Subject string `json:"subject"`
	}
	var got request

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
			t.Errorf("content-type = %q", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	if err := New(nil).PostJSON(context.Background(), srv.URL, request{Subject: "hi"}, nil); err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
	if got.Subject != "hi" {
		t.Errorf("subject = %q, want %q", got.Subject, "hi")
	}
}

func TestPostJSON_DecodesResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"code":0,"msg":"ok"}`)
	}))
	t.Cleanup(srv.Close)

	var out struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := New(nil).PostJSON(context.Background(), srv.URL, nil, &out); err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
	if out.Code != 0 || out.Msg != "ok" {
		t.Errorf("out = %+v", out)
	}
}

func TestPostJSON_NonSuccessReturnsError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `boom`)
	}))
	t.Cleanup(srv.Close)

	err := New(nil).PostJSON(context.Background(), srv.URL, nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "http 500") || !strings.Contains(err.Error(), "boom") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostJSON_AcceptsRawBytes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "raw-bytes" {
			t.Errorf("body = %q, want %q", body, "raw-bytes")
		}
	}))
	t.Cleanup(srv.Close)

	if err := New(nil).PostJSON(context.Background(), srv.URL, []byte("raw-bytes"), nil); err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
}

func TestPostJSON_RespectsCancelledContext(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := New(nil).PostJSON(ctx, srv.URL, nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestPostJSON_RedactsURLOnNetworkError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	srv.Close()

	const secret = "secret-token-1eaf"
	target := srv.URL + "/bot/" + secret + "?token=" + secret

	err := New(nil).PostJSON(context.Background(), target, nil, nil)
	if err == nil {
		t.Fatal("expected error after server closed, got nil")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("error string leaks credential %q: %v", secret, err)
	}

	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		t.Fatalf("err does not unwrap to *url.Error: %v", err)
	}
	if strings.Contains(urlErr.URL, secret) {
		t.Errorf("url.Error.URL leaks credential %q: %q", secret, urlErr.URL)
	}
}

func TestPostJSON_RedactionPreservesContextSentinel(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := New(nil).PostJSON(ctx, srv.URL+"?token=abc", nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want errors.Is(..., context.Canceled)", err)
	}
}

func TestPostJSON_InvalidJSONResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `not json`)
	}))
	t.Cleanup(srv.Close)

	var out struct {
		Code int `json:"code"`
	}
	err := New(nil).PostJSON(context.Background(), srv.URL, nil, &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal response") {
		t.Errorf("unexpected error: %v", err)
	}
}
