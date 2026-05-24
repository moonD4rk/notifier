// Basic broadcast example: send a single message to every configured
// channel concurrently.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/moond4rk/notifier"
	"github.com/moond4rk/notifier/provider/bark"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	n := notifier.New(
		notifier.WithDingTalk(os.Getenv("DINGTALK_TOKEN"), os.Getenv("DINGTALK_SECRET")),
		notifier.WithBark(os.Getenv("BARK_KEY"), bark.DefaultServer),
		notifier.WithFeishu(os.Getenv("FEISHU_TOKEN"), os.Getenv("FEISHU_SECRET")),
		notifier.WithLark(os.Getenv("LARK_TOKEN"), os.Getenv("LARK_SECRET")),
		notifier.WithServerChan("", os.Getenv("SERVER_CHAN_SEND_KEY")),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := n.Send(ctx, notifier.Message{
		Subject: "this is subject",
		Content: "this is content",
		Extras:  map[string]any{bark.KeySound: "siren.caf"},
	})
	if err == nil {
		return nil
	}

	var multi *notifier.MultiError
	if errors.As(err, &multi) {
		for _, e := range multi.Errors {
			log.Printf("provider %s failed: %v", e.Provider, e.Err)
		}
	}
	return err
}
