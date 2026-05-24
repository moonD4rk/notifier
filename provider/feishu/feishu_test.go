package feishu

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	if p := New(Config{Token: "t"}); p.Name() != "feishu" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestSend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		respStatus int
		respBody   string
		wantErr    bool
		errContain string
	}{
		{name: "ok", respStatus: http.StatusOK, respBody: `{"StatusMessage":"success"}`},
		{name: "body error", respStatus: http.StatusOK, respBody: `{"StatusMessage":"forbidden"}`, wantErr: true, errContain: "forbidden"},
		{name: "http error", respStatus: http.StatusInternalServerError, respBody: `boom`, wantErr: true, errContain: "http 500"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var bodyJSON map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Path; got != "/feishu-tok" {
					t.Errorf("path = %q", got)
				}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &bodyJSON)
				w.WriteHeader(tc.respStatus)
				_, _ = io.WriteString(w, tc.respBody)
			}))
			t.Cleanup(srv.Close)

			p := New(Config{Token: "feishu-tok"})
			p.endpoint = srv.URL
			p.now = func() time.Time { return time.Unix(1700000000, 0) }

			err := p.Send(context.Background(), provider.Message{Subject: "subj", Content: "body"})
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.errContain != "" && !strings.Contains(err.Error(), tc.errContain) {
				t.Errorf("err %v does not contain %q", err, tc.errContain)
			}
			if bodyJSON["msg_type"] != "post" {
				t.Errorf("msg_type = %v", bodyJSON["msg_type"])
			}
		})
	}
}
