package bark

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moond4rk/notifier/provider"
)

func TestNew_EmptyKeyReturnsNil(t *testing.T) {
	t.Parallel()
	if got := New(Config{}); got != nil {
		t.Errorf("New with empty key returned %v, want nil", got)
	}
}

func TestNew_DefaultEndpoint(t *testing.T) {
	t.Parallel()
	p := New(Config{Key: "device"})
	if p == nil {
		t.Fatal("nil provider")
	}
	if want := "https://" + DefaultServer; p.endpoint != want {
		t.Errorf("endpoint = %q, want %q", p.endpoint, want)
	}
}

func TestNew_CustomServer(t *testing.T) {
	t.Parallel()
	p := New(Config{Key: "device", Server: "bark.example.com"})
	if p.endpoint != "https://bark.example.com" {
		t.Errorf("endpoint = %q", p.endpoint)
	}
}

func TestProvider_Name(t *testing.T) {
	t.Parallel()
	p := New(Config{Key: "k"})
	if p.Name() != "bark" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestSend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		msg        provider.Message
		respStatus int
		wantErr    bool
		wantSound  string
		wantIcon   string
	}{
		{
			name:       "ok default sound",
			msg:        provider.Message{Subject: "subj", Content: "body"},
			respStatus: http.StatusOK,
			wantSound:  "alarm.caf",
		},
		{
			name: "extras override sound and icon",
			msg: provider.Message{
				Subject: "subj", Content: "body",
				Extras: map[string]any{
					KeySound: "siren.caf",
					KeyIcon:  "https://i/icon.png",
					KeyGroup: "alerts",
					KeyURL:   "https://example.com",
					KeyBadge: 7,
				},
			},
			respStatus: http.StatusOK,
			wantSound:  "siren.caf",
			wantIcon:   "https://i/icon.png",
		},
		{
			name:       "5xx",
			msg:        provider.Message{Subject: "s"},
			respStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got payload
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/push" {
					t.Errorf("path = %q", r.URL.Path)
				}
				body, _ := io.ReadAll(r.Body)
				if err := json.Unmarshal(body, &got); err != nil {
					t.Errorf("unmarshal: %v", err)
				}
				w.WriteHeader(tc.respStatus)
			}))
			t.Cleanup(srv.Close)

			p := New(Config{Key: "device-key"})
			p.endpoint = srv.URL

			err := p.Send(context.Background(), tc.msg)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got.DeviceKey != "device-key" {
				t.Errorf("device_key = %q", got.DeviceKey)
			}
			if got.Sound != tc.wantSound {
				t.Errorf("sound = %q, want %q", got.Sound, tc.wantSound)
			}
			if tc.wantIcon != "" && got.Icon != tc.wantIcon {
				t.Errorf("icon = %q, want %q", got.Icon, tc.wantIcon)
			}
		})
	}
}
