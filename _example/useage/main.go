package main

import (
	"github.com/moond4rk/notifier"
	"github.com/moond4rk/notifier/provider/bark"
)

func main() {
	var (
		dingtalkToken  = "dingtalk_token"
		dingtalkSecret = "dingtalk_secret"
		barkKey        = "bark_key"
		barkServer     = bark.DefaultBarkServer
	)

	notifier := notifier.New(
		notifier.WithDingTalk(dingtalkToken, dingtalkSecret),
		notifier.WithBark(barkKey, barkServer),
	)

	var (
		subject = "this is subject"
		content = "this is content"
	)
	if err := notifier.Send(subject, content); err != nil {
		panic(err)
	}
}
