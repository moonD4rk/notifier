# notifier
notifier is a simple Go library to send notification to other applications.

## Feature

| Provider                                                     | Code |
| ------------------------------------------------------------ | ---- |
| [DingTalk](https://www.dingtalk.com/en)                      |      |
| [Bark](https://apps.apple.com/us/app/bark-customed-notifications/id1403753865) |      |

## Install

`go get -u github.com/moond4rk/notifier`

## Useage



```go
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


```

