package main

import (
	"os"

	"github.com/moond4rk/notifier"
)

func main() {
	var (
		dingtalkToken  = os.Getenv("dingtalk_token")
		dingtalkSecret = os.Getenv("dingtalk_secret")
		barkKey        = os.Getenv("bark_key")
		barkServer     = notifier.DefaultBarkServer
		feishuToken    = os.Getenv("feishu_token")
		feishuSecret   = os.Getenv("feishu_secret")
		larkToken      = os.Getenv("feishu_token")
		larkSecret     = os.Getenv("feishu_secret")
	)
	notifier := notifier.New(
		notifier.WithDingTalk(dingtalkToken, dingtalkSecret),
		notifier.WithBark(barkKey, barkServer),
		notifier.WithFeishu(feishuToken, feishuSecret),
		notifier.WithLark(larkToken, larkSecret),
	)

	var (
		subject = "this is subject"
		content = "this is content"
	)
	if err := notifier.Send(subject, content); err != nil {
		panic(err)
	}
}
