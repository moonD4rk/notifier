package serverchan

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moond4rk/notifier/provider"
)

func TestNew_EmptySendKeyReturnsNil(t *testing.T) {
	t.Parallel()
	if got := New(Config{UserID: "uid"}); got != nil {
		t.Errorf("New with empty sendkey returned %v, want nil", got)
	}
}

func TestProvider_Name(t *testing.T) {
	t.Parallel()
	if p := New(Config{SendKey: "k"}); p.Name() != "serverchan" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestSend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		respStatus int
		wantErr    bool
	}{
		{name: "ok", respStatus: http.StatusOK},
		{name: "5xx", respStatus: http.StatusInternalServerError, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got payload
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/sendkey-abc.send" {
					t.Errorf("path = %q", r.URL.Path)
				}
				body, _ := io.ReadAll(r.Body)
				if err := json.Unmarshal(body, &got); err != nil {
					t.Errorf("unmarshal: %v", err)
				}
				w.WriteHeader(tc.respStatus)
			}))
			t.Cleanup(srv.Close)

			p := New(Config{UserID: "uid", SendKey: "sendkey-abc"})
			p.endpoint = srv.URL

			err := p.Send(context.Background(), provider.Message{Subject: "subj", Content: "body"})
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if got.Text != "subj" || got.Desp != "body" {
				t.Errorf("payload = %+v", got)
			}
		})
	}
}
