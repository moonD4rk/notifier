package crypto

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLarkSign_Deterministic(t *testing.T) {
	t.Parallel()

	a := LarkSign(1700000000, "secret")
	b := LarkSign(1700000000, "secret")
	if a != b {
		t.Fatalf("non-deterministic")
	}
}

func TestLarkSign_VariesWithInputs(t *testing.T) {
	t.Parallel()

	base := LarkSign(1700000000, "secret")
	if got := LarkSign(1700000001, "secret"); got == base {
		t.Errorf("different timestamps must yield different signatures")
	}
	if got := LarkSign(1700000000, "other"); got == base {
		t.Errorf("different secrets must yield different signatures")
	}
}

func TestLarkPayload_NoSecretOmitsSign(t *testing.T) {
	t.Parallel()

	raw, err := LarkPayload("hello", "world", "", time.Unix(1700000000, 0))
	if err != nil {
		t.Fatalf("LarkPayload: %v", err)
	}
	var pd map[string]any
	if err := json.Unmarshal(raw, &pd); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if pd["msg_type"] != "post" {
		t.Errorf("msg_type = %v, want post", pd["msg_type"])
	}
	if pd["timestamp"] != "1700000000" {
		t.Errorf("timestamp = %v", pd["timestamp"])
	}
	if _, ok := pd["sign"]; ok {
		t.Error("sign should be omitted when secret is empty")
	}
}

func TestLarkPayload_WithSecretIncludesSign(t *testing.T) {
	t.Parallel()

	raw, err := LarkPayload("hi", "body", "topsecret", time.Unix(1700000000, 0))
	if err != nil {
		t.Fatalf("LarkPayload: %v", err)
	}
	var pd struct {
		Timestamp string `json:"timestamp"`
		Sign      string `json:"sign"`
		MsgType   string `json:"msg_type"`
	}
	if err := json.Unmarshal(raw, &pd); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if pd.Timestamp != "1700000000" {
		t.Errorf("timestamp = %q", pd.Timestamp)
	}
	if pd.Sign == "" {
		t.Error("sign must be set when secret is non-empty")
	}
	if want := LarkSign(1700000000, "topsecret"); pd.Sign != want {
		t.Errorf("sign = %q, want %q", pd.Sign, want)
	}
	if pd.MsgType != "post" {
		t.Errorf("msg_type = %q", pd.MsgType)
	}
}

func TestLarkPayload_EncodesSubjectAndContent(t *testing.T) {
	t.Parallel()

	raw, err := LarkPayload("alarm", "fire in srv-01", "", time.Unix(0, 0))
	if err != nil {
		t.Fatalf("LarkPayload: %v", err)
	}
	var pd struct {
		Content struct {
			Post struct {
				ZhCN struct {
					Title   string             `json:"title"`
					Content [][]larkContentMsg `json:"content"`
				} `json:"zh_cn"`
			} `json:"post"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &pd); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if pd.Content.Post.ZhCN.Title != "alarm" {
		t.Errorf("title = %q", pd.Content.Post.ZhCN.Title)
	}
	if len(pd.Content.Post.ZhCN.Content) != 1 || len(pd.Content.Post.ZhCN.Content[0]) != 1 {
		t.Fatalf("unexpected content shape: %+v", pd.Content.Post.ZhCN.Content)
	}
	msg := pd.Content.Post.ZhCN.Content[0][0]
	if msg.Tag != "text" || msg.Text != "fire in srv-01" {
		t.Errorf("msg = %+v", msg)
	}
}
