package notifier_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/moond4rk/notifier"
	"github.com/moond4rk/notifier/provider"
)

// fakeProvider is an in-memory Provider used by these tests.
type fakeProvider struct {
	name string
	err  error
	hits atomic.Int32
	mu   sync.Mutex
	last provider.Message
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) Send(_ context.Context, msg provider.Message) error {
	f.hits.Add(1)
	f.mu.Lock()
	f.last = msg
	f.mu.Unlock()
	return f.err
}

func TestSend_NoProvider(t *testing.T) {
	t.Parallel()

	n := notifier.New()
	err := n.Send(context.Background(), notifier.Message{Subject: "x"})
	if !errors.Is(err, notifier.ErrNoProvider) {
		t.Errorf("err = %v, want ErrNoProvider", err)
	}
}

func TestSend_AllSucceed(t *testing.T) {
	t.Parallel()

	a := &fakeProvider{name: "a"}
	b := &fakeProvider{name: "b"}
	n := notifier.New(notifier.WithProvider(a), notifier.WithProvider(b))

	if err := n.Send(context.Background(), notifier.Message{Subject: "hi"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if a.hits.Load() != 1 || b.hits.Load() != 1 {
		t.Errorf("hits a=%d b=%d", a.hits.Load(), b.hits.Load())
	}
}

func TestSend_AggregatesFailures(t *testing.T) {
	t.Parallel()

	bad := &fakeProvider{name: "bad", err: errors.New("boom")}
	good := &fakeProvider{name: "good"}
	n := notifier.New(notifier.WithProvider(bad), notifier.WithProvider(good))

	err := n.Send(context.Background(), notifier.Message{Subject: "hi"})
	if err == nil {
		t.Fatal("expected error")
	}
	var multi *notifier.MultiError
	if !errors.As(err, &multi) {
		t.Fatalf("err is not *MultiError: %T (%v)", err, err)
	}
	if len(multi.Errors) != 1 {
		t.Fatalf("len(errors) = %d, want 1", len(multi.Errors))
	}
	if multi.Errors[0].Provider != "bad" {
		t.Errorf("provider = %q", multi.Errors[0].Provider)
	}
	if good.hits.Load() != 1 {
		t.Errorf("good provider should still run; hits=%d", good.hits.Load())
	}
}

func TestSend_PassesContext(t *testing.T) {
	t.Parallel()

	a := &fakeProvider{name: "a"}
	n := notifier.New(notifier.WithProvider(a))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := n.Send(ctx, notifier.Message{Subject: "hi"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestWithProvider_NilIgnored(t *testing.T) {
	t.Parallel()

	n := notifier.New(notifier.WithProvider(nil))
	err := n.Send(context.Background(), notifier.Message{})
	if !errors.Is(err, notifier.ErrNoProvider) {
		t.Errorf("err = %v, want ErrNoProvider", err)
	}
}

func TestConvenienceHelpers_NoCredentialsAreNoOp(t *testing.T) {
	t.Parallel()

	n := notifier.New(
		notifier.WithDingTalk("", ""),
		notifier.WithBark("", ""),
		notifier.WithLark("", ""),
		notifier.WithFeishu("", ""),
		notifier.WithServerChan("", ""),
		notifier.WithWeCom(""),
		notifier.WithTelegram("", ""),
	)
	if got := len(n.Providers()); got != 0 {
		t.Errorf("len(providers) = %d, want 0", got)
	}
}

func TestMultiError_UnwrapAllowsErrorsIs(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("sentinel")
	bad := &fakeProvider{name: "bad", err: sentinel}
	n := notifier.New(notifier.WithProvider(bad))

	err := n.Send(context.Background(), notifier.Message{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("errors.Is(err, sentinel) = false; err = %v", err)
	}
}
