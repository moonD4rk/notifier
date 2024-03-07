package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Telegram struct holds necessary data to communicate with the Telegram API.
type Telegram struct {
	client  *tgbotapi.BotAPI
	chatIDs []int64
}

// New returns a new instance of a Telegram notification service.
// For more information about telegram api token:
//
//	-> https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api#NewBotAPI
func New(apiToken string, chatIDs []int64) (*Telegram, error) {
	client, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram client: %w", err)
	}

	t := &Telegram{
		client:  client,
		chatIDs: chatIDs,
	}
	if len(t.chatIDs) == 0 {
		return nil, fmt.Errorf("chatIDs is empty")
	}

	return t, nil
}

// Send sends a message to the chat IDs added to the internal chat ID list.
func (t *Telegram) Send(ctx context.Context, subject, message string) error {
	fullMessage := subject + "\n" + message

	msg := tgbotapi.NewMessage(0, fullMessage)
	msg.ParseMode = tgbotapi.ModeMarkdown

	for _, chatID := range t.chatIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg.ChatID = chatID
			_, err := t.client.Send(msg)
			if err != nil {
				return fmt.Errorf("failed to send message to Telegram chat %q: %w", chatID, err)
			}
		}
	}

	return nil
}
