package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestDingTalkSign_Deterministic(t *testing.T) {
	t.Parallel()

	a := DingTalkSign(1700000000000, "secret")
	b := DingTalkSign(1700000000000, "secret")
	if a != b {
		t.Fatalf("non-deterministic: %q vs %q", a, b)
	}
}

func TestDingTalkSign_VariesWithInputs(t *testing.T) {
	t.Parallel()

	base := DingTalkSign(1700000000000, "secret")
	if got := DingTalkSign(1700000000001, "secret"); got == base {
		t.Errorf("different timestamps must yield different signatures")
	}
	if got := DingTalkSign(1700000000000, "other"); got == base {
		t.Errorf("different secrets must yield different signatures")
	}
}

func TestDingTalkSign_MatchesReferenceImplementation(t *testing.T) {
	t.Parallel()

	const (
		timestamp = int64(1700000000000)
		secret    = "topsecret"
	)
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = fmt.Fprintf(h, "%d\n%s", timestamp, secret)
	want := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if got := DingTalkSign(timestamp, secret); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDingTalkSign_DecodesToSHA256Digest(t *testing.T) {
	t.Parallel()

	sig := DingTalkSign(1700000000000, "secret")
	raw, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		t.Fatalf("invalid base64: %v", err)
	}
	if len(raw) != sha256.Size {
		t.Errorf("len(raw) = %d, want %d", len(raw), sha256.Size)
	}
}
