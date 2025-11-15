package bot

import (
	"context"
	"fmt"

	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"first-max-bot/internal/state"
)

type Request struct {
	Context   context.Context
	Update    *schemes.MessageCreatedUpdate
	Command   string
	Args      string
	UserState *state.UserState
	Metadata  map[string]any
}

func (r *Request) Recipient() schemes.Recipient {
	if r.Update != nil {
		return r.Update.Message.Recipient
	}
	// Для callback'ов recipient хранится в Metadata
	if r.Metadata != nil {
		if recipient, ok := r.Metadata["recipient"].(schemes.Recipient); ok {
			return recipient
		}
	}
	return schemes.Recipient{}
}

func (r *Request) Sender() schemes.User {
	if r.Update != nil {
		return r.Update.Message.Sender
	}
	if r.Metadata != nil {
		if sender, ok := r.Metadata["sender"].(schemes.User); ok {
			return sender
		}
	}
	return schemes.User{}
}

func (r *Request) Text() string {
	if r.Update != nil {
		return r.Update.Message.Body.Text
	}
	if r.Metadata != nil {
		if text, ok := r.Metadata["text"].(string); ok {
			return text
		}
	}
	return ""
}

func (r *Request) UserID() string {
	if r.Update != nil {
		if id := r.Update.Message.Sender.UserId; id != 0 {
			return fmt.Sprintf("%d", id)
		}
		if id := r.Update.Message.Recipient.UserId; id != 0 {
			return fmt.Sprintf("%d", id)
		}
	}
	if r.Metadata != nil {
		if senderID, ok := r.Metadata["sender_id"].(string); ok {
			return senderID
		}
	}
	return ""
}

func (r *Request) MessageID() string {
	if r.Update != nil && r.Update.Message.Body.Mid != "" {
		return r.Update.Message.Body.Mid
	}
	if r.Metadata != nil {
		if messageID, ok := r.Metadata["message_id"].(string); ok {
			return messageID
		}
	}
	return ""
}
