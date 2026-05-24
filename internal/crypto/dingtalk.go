package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// DingTalkSign returns the HMAC-SHA256 signature used by signed DingTalk
// webhooks. timestampMs is Unix milliseconds; secret is the bot secret.
//
// See https://open.dingtalk.com/document/robots/customize-robot-security-settings
// for the format.
func DingTalkSign(timestampMs int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestampMs, secret)
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
