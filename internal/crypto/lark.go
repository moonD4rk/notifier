package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// LarkSign returns the signature used by signed Lark/Feishu webhooks.
// timestampSec is Unix seconds; secret is the bot secret.
//
// See https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/bot-v2/use-custom-bots-in-a-group/signature-verification
// for the format.
func LarkSign(timestampSec int64, secret string) string {
	key := fmt.Sprintf("%d\n%s", timestampSec, secret)
	h := hmac.New(sha256.New, []byte(key))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// LarkPayload builds the JSON body for a Lark/Feishu webhook message. If
// secret is non-empty, timestamp + sign fields are included for signature
// verification.
func LarkPayload(subject, content, secret string, now time.Time) ([]byte, error) {
	timestamp := strconv.FormatInt(now.Unix(), 10)
	var sign string
	if secret != "" {
		sign = LarkSign(now.Unix(), secret)
	}
	pd := &larkPostData{
		Timestamp: timestamp,
		Sign:      sign,
		MsgType:   "post",
	}
	pd.Content.Post.ZhCN.Title = subject
	pd.Content.Post.ZhCN.Content = append(
		pd.Content.Post.ZhCN.Content,
		[]larkContentMsg{{Tag: "text", Text: content}},
	)
	return json.Marshal(pd)
}

type larkPostData struct {
	Timestamp string      `json:"timestamp"`
	Sign      string      `json:"sign,omitempty"`
	MsgType   string      `json:"msg_type"`
	Content   larkContent `json:"content"`
}

type larkContent struct {
	Post larkPost `json:"post"`
}

type larkPost struct {
	ZhCN larkZhCN `json:"zh_cn,omitempty"`
}

type larkZhCN struct {
	Title   string             `json:"title"`
	Content [][]larkContentMsg `json:"content"`
}

type larkContentMsg struct {
	Tag  string `json:"tag"`
	Text string `json:"text"`
}
