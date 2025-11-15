package bot

import (
	"context"

	"github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

type Responder interface {
	SendText(ctx context.Context, recipient schemes.Recipient, text string) error
	SendMarkdown(ctx context.Context, recipient schemes.Recipient, text string) error
	SendMessage(ctx context.Context, message *maxbot.Message) error
	SendTextWithKeyboard(ctx context.Context, recipient schemes.Recipient, text string, keyboard *maxbot.Keyboard) error
	SendMarkdownWithKeyboard(ctx context.Context, recipient schemes.Recipient, text string, keyboard *maxbot.Keyboard) error
	SendTextWithFile(ctx context.Context, recipient schemes.Recipient, text string, fileToken string) error
	AnswerCallback(ctx context.Context, callbackID string, answer *schemes.CallbackAnswer) error
	AnswerCallbackWithEdit(ctx context.Context, callbackID string, text string, keyboard *maxbot.Keyboard) error
	DeleteMessageBySeq(ctx context.Context, messageSeq int64) error
	DeleteMessageByMid(ctx context.Context, messageID string) error
	NewKeyboardBuilder() *maxbot.Keyboard
}
