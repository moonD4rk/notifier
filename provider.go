package notifier

type provider interface {
	Send(subject, content string) error
}
