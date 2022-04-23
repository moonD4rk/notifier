# notifier
notifier is a simple Go library to send notification to other applications.

## Feature

| Provider                                                     | Code                                                         |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| [DingTalk](https://www.dingtalk.com/en)                      | [provider/bark](https://github.com/moonD4rk/notifier/tree/main/provider/bark) |
| [Bark](https://apps.apple.com/us/app/bark-customed-notifications/id1403753865) | [provider/dingtalk](https://github.com/moonD4rk/notifier/tree/main/provider/dingtalk) |

## Install

`go get -u github.com/moond4rk/notifier`

## Usage



```go
package main

import (
	"github.com/moond4rk/notifier"
)

func main() {
	var (
		dingtalkToken  = "dingtalk_token"
		dingtalkSecret = "dingtalk_secret"
		barkKey        = "bark_key"
		barkServer     = notifier.DefaultBarkServer
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

<img src="https://raw.githubusercontent.com/moonD4rk/staticfiles/master/picture/notifier-screenshot.png" width="480" align="left"/>
