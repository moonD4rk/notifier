package notifier

type provider interface {
	Send(title, content string) error
}
