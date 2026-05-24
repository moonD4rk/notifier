// Package notifier dispatches a single message to multiple notification
// channels concurrently.
//
// A Notifier is constructed with one or more provider options and then
// used to broadcast a Message:
//
//	n := notifier.New(
//	    notifier.WithDingTalk(os.Getenv("DINGTALK_TOKEN"), os.Getenv("DINGTALK_SECRET")),
//	    notifier.WithBark(os.Getenv("BARK_KEY"), ""),
//	)
//	if err := n.Send(ctx, notifier.Message{
//	    Subject: "alert",
//	    Content: "the disk is full",
//	}); err != nil {
//	    log.Print(err)
//	}
//
// Each provider runs in its own goroutine; failures are aggregated into
// a *MultiError so a single broken channel does not silence the others.
//
// Channel-specific options (Bark sound, Lark markdown, ...) travel via
// Message.Extras using string keys defined in the corresponding provider
// package.
//
// # Custom providers
//
// Anything that implements provider.Provider can be registered with
// WithProvider. The provider package documents the contract.
package notifier
