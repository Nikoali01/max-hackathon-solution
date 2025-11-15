package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"first-max-bot/internal/state"
)

type Bot struct {
	api    *maxbot.Api
	router *Router
	state  state.Repository
	logger zerolog.Logger
}

func New(api *maxbot.Api, router *Router, stateRepo state.Repository, logger zerolog.Logger) *Bot {
	return &Bot{
		api:    api,
		router: router,
		state:  stateRepo,
		logger: logger,
	}
}

func (b *Bot) Run(ctx context.Context) error {
	updates := b.api.GetUpdates(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case upd, ok := <-updates:
			if !ok {
				return nil
			}
			b.handleUpdate(ctx, upd)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update schemes.UpdateInterface) {
	switch upd := update.(type) {
	case *schemes.MessageCreatedUpdate:
		b.handleMessage(ctx, upd)
	case *schemes.MessageCallbackUpdate:
		b.handleCallback(ctx, upd)
	default:
		b.logger.Debug().Str("type", fmt.Sprintf("%T", update)).Msg("update ignored")
	}
}

func (b *Bot) handleMessage(ctx context.Context, upd *schemes.MessageCreatedUpdate) {
	logger := b.logger.With().
		Int64("chat_id", upd.Message.Recipient.ChatId).
		Int64("user_id", upd.Message.Sender.UserId).
		Str("text", upd.Message.Body.Text).
		Logger()

	// Сначала загружаем состояние, чтобы проверить, не в процессе ли регистрации
	var (
		userID    = extractUserID(upd)
		userState *state.UserState
		err       error
	)
	if userID != "" {
		userState, err = b.state.GetUserState(ctx, userID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to load user state")
		}
	}

	if userState == nil {
		userState = &state.UserState{
			UserRegistrationData: make(map[string]string),
		}
	}
	if userState.UserRegistrationData == nil {
		userState.UserRegistrationData = make(map[string]string)
	}

	// Разрешаем handler с учетом состояния
	handler, command, args := b.router.ResolveByState(upd.Message.Body.Text, userState)
	if handler == nil {
		logger.Warn().Msg("no handler registered")
		return
	}

	req := &Request{
		Context:   ctx,
		Update:    upd,
		Command:   command,
		Args:      args,
		UserState: userState,
	}

	if err := handler.Handle(ctx, req, b); err != nil {
		logger.Error().Err(err).Msg("handler failed")
	}

	if userID != "" {
		stateToSave := state.UserState{
			LastCommand: command,
			LastUpdated: time.Now(),
		}
		if req.UserState != nil {
			stateToSave.UserRegistrationStep = req.UserState.UserRegistrationStep
			stateToSave.UserRegistrationData = req.UserState.UserRegistrationData
		}
		if stateToSave.UserRegistrationData == nil {
			stateToSave.UserRegistrationData = make(map[string]string)
		}
		if err := b.state.SaveUserState(ctx, userID, stateToSave); err != nil {
			logger.Error().Err(err).Msg("failed to save user state")
		}
	}
}

func extractUserID(upd *schemes.MessageCreatedUpdate) string {
	if upd.Message.Sender.UserId != 0 {
		return fmt.Sprintf("%d", upd.Message.Sender.UserId)
	}
	if upd.Message.Recipient.UserId != 0 {
		return fmt.Sprintf("%d", upd.Message.Recipient.UserId)
	}
	return ""
}

func (b *Bot) SendText(ctx context.Context, recipient schemes.Recipient, text string) error {
	message := maxbot.NewMessage()
	if recipient.ChatId != 0 {
		message.SetChat(recipient.ChatId)
	}
	if recipient.UserId != 0 {
		message.SetUser(recipient.UserId)
	}
	message.SetText(text)
	_, err := b.api.Messages.Send(ctx, message)
	a, _ := err.(schemes.Error)
	if a.Code == "" {
		return nil
	}
	return err
}

func (b *Bot) SendMarkdown(ctx context.Context, recipient schemes.Recipient, text string) error {
	message := maxbot.NewMessage()
	if recipient.ChatId != 0 {
		message.SetChat(recipient.ChatId)
	}
	if recipient.UserId != 0 {
		message.SetUser(recipient.UserId)
	}
	message.SetText(text)
	message.SetFormat("markdown")
	_, err := b.api.Messages.Send(ctx, message)
	a, _ := err.(schemes.Error)
	if a.Code == "" {
		return nil
	}
	return err
}

func (b *Bot) SendMessage(ctx context.Context, message *maxbot.Message) error {
	_, err := b.api.Messages.Send(ctx, message)
	return err
}

func (b *Bot) AnswerCallback(ctx context.Context, callbackID string, answer *schemes.CallbackAnswer) error {
	_, err := b.api.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

// AnswerCallbackWithEdit редактирует сообщение через CallbackAnswer
// Это правильный способ редактировать сообщение при callback - через CallbackAnswer.Message
func (b *Bot) AnswerCallbackWithEdit(ctx context.Context, callbackID string, text string, keyboard *maxbot.Keyboard) error {
	answer := &schemes.CallbackAnswer{}

	// Создаем новое сообщение для редактирования
	messageBody := &schemes.NewMessageBody{
		Text: text,
	}

	// Если клавиатура передана, добавляем её в attachments
	// Если keyboard == nil, attachments останется nil/пустым, что удалит клавиатуру из сообщения
	if keyboard != nil {
		messageBody.Attachments = []interface{}{
			schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build()),
		}
	}
	// Если keyboard == nil, Attachments будет nil (благодаря omitempty в JSON теге),
	// что должно удалить существующую клавиатуру

	answer.Message = messageBody

	_, err := b.api.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

// DeleteMessageBySeq удаляет сообщение по sequence number (Seq)
func (b *Bot) DeleteMessageBySeq(ctx context.Context, messageSeq int64) error {
	if messageSeq == 0 {
		return fmt.Errorf("message sequence is required for delete")
	}

	b.logger.Debug().Int64("seq", messageSeq).Msg("attempting to delete message")

	result, err := b.api.Messages.DeleteMessage(ctx, messageSeq)
	if err != nil {
		b.logger.Error().Err(err).Int64("seq", messageSeq).Msg("delete message failed")
		return fmt.Errorf("delete message failed: %w", err)
	}

	// Проверяем результат
	if result != nil {
		if !result.Success {
			b.logger.Warn().Str("message", result.Message).Int64("seq", messageSeq).Msg("delete returned success=false")
			return fmt.Errorf("delete failed: %s", result.Message)
		}
		b.logger.Debug().Int64("seq", messageSeq).Msg("message deleted successfully")
	}

	return nil
}

// DeleteMessageByMid удаляет сообщение по Mid (строка)
// Использует новый метод DeleteMessageByStringID из локальной библиотеки
func (b *Bot) DeleteMessageByMid(ctx context.Context, messageID string) error {
	if messageID == "" {
		return fmt.Errorf("message ID is required for delete")
	}

	b.logger.Debug().Str("message_id", messageID).Msg("attempting to delete message by Mid")

	result, err := b.api.Messages.DeleteMessageByStringID(ctx, messageID)
	if err != nil {
		b.logger.Error().Err(err).Str("message_id", messageID).Msg("delete message failed")
		return fmt.Errorf("delete message failed: %w", err)
	}

	// Проверяем результат
	if result != nil {
		if !result.Success {
			b.logger.Warn().Str("message", result.Message).Str("message_id", messageID).Msg("delete returned success=false")
			return fmt.Errorf("delete failed: %s", result.Message)
		}
		b.logger.Debug().Str("message_id", messageID).Msg("message deleted successfully by Mid")
	}

	return nil
}

func (b *Bot) NewKeyboardBuilder() *maxbot.Keyboard {
	return b.api.Messages.NewKeyboardBuilder()
}

func (b *Bot) SendTextWithKeyboard(ctx context.Context, recipient schemes.Recipient, text string, keyboard *maxbot.Keyboard) error {
	message := maxbot.NewMessage()
	if recipient.ChatId != 0 {
		message.SetChat(recipient.ChatId)
	}
	if recipient.UserId != 0 {
		message.SetUser(recipient.UserId)
	}
	message.SetText(text)
	if keyboard != nil {
		message.AddKeyboard(keyboard)
	}
	_, err := b.api.Messages.Send(ctx, message)
	a, _ := err.(schemes.Error)
	if a.Code == "" {
		return nil
	}
	return err
}

func (b *Bot) SendMarkdownWithKeyboard(ctx context.Context, recipient schemes.Recipient, text string, keyboard *maxbot.Keyboard) error {
	message := maxbot.NewMessage()
	if recipient.ChatId != 0 {
		message.SetChat(recipient.ChatId)
	}
	if recipient.UserId != 0 {
		message.SetUser(recipient.UserId)
	}
	message.SetText(text)
	message.SetFormat("markdown")
	if keyboard != nil {
		message.AddKeyboard(keyboard)
	}
	_, err := b.api.Messages.Send(ctx, message)
	a, _ := err.(schemes.Error)
	if a.Code == "" {
		return nil
	}
	return err
}

func (b *Bot) SendTextWithFile(ctx context.Context, recipient schemes.Recipient, text string, fileToken string) error {
	if fileToken == "" {
		// Если токен пустой, отправляем только текст
		return b.SendText(ctx, recipient, text)
	}

	// Создаем UploadedInfo из token
	uploadedInfo := schemes.UploadedInfo{
		Token: fileToken,
	}

	// Создаем сообщение с файлом
	message := maxbot.NewMessage()
	if recipient.ChatId != 0 {
		message.SetChat(recipient.ChatId)
	}
	if recipient.UserId != 0 {
		message.SetUser(recipient.UserId)
	}
	message.SetText(text)
	message.AddFile(&uploadedInfo)

	_, err := b.api.Messages.Send(ctx, message)
	a, _ := err.(schemes.Error)
	if a.Code == "" {
		return nil
	}
	return err
}

func (b *Bot) handleCallback(ctx context.Context, upd *schemes.MessageCallbackUpdate) {
	logger := b.logger.With().
		Int64("user_id", upd.Callback.User.UserId).
		Str("payload", upd.Callback.Payload).
		Logger()

	userID := fmt.Sprintf("%d", upd.Callback.User.UserId)
	userState, err := b.state.GetUserState(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to load user state")
		userState = &state.UserState{
			UserRegistrationData: make(map[string]string),
		}
	}
	if userState == nil {
		userState = &state.UserState{
			UserRegistrationData: make(map[string]string),
		}
	}
	if userState.UserRegistrationData == nil {
		userState.UserRegistrationData = make(map[string]string)
	}

	// Определяем recipient из callback
	recipient := schemes.Recipient{
		UserId:   upd.Callback.User.UserId,
		ChatType: schemes.DIALOG,
	}
	if upd.Message != nil {
		recipient.ChatId = upd.Message.Recipient.ChatId
		recipient.ChatType = upd.Message.Recipient.ChatType
	}

	// Ищем handler по registration step или по payload
	handler := b.router.ResolveCallback(upd.Callback.Payload, userState)
	if handler == nil {
		logger.Warn().Str("payload", upd.Callback.Payload).Int64("user_id", upd.Callback.User.UserId).Msg("no callback handler found")
		// Отвечаем на callback чтобы убрать loading
		b.api.Messages.AnswerOnCallback(ctx, upd.Callback.CallbackID, &schemes.CallbackAnswer{
			Notification: "Команда не распознана",
		})
		return
	}

	// Получаем message ID (Mid) для удаления - API требует строку, а не число!
	messageID := ""
	if upd.Message != nil && upd.Message.Body.Mid != "" {
		messageID = upd.Message.Body.Mid
	}

	messageSeq := int64(0)
	if upd.Message != nil && upd.Message.Body.Seq != 0 {
		messageSeq = upd.Message.Body.Seq
	}

	// Создаем специальный request для callback
	req := &Request{
		Context:   ctx,
		Update:    nil, // для callback нет MessageCreatedUpdate
		Command:   "",  // callback не является командой
		Args:      upd.Callback.Payload,
		UserState: userState, // передаем указатель, чтобы handler мог его изменять
		Metadata: map[string]any{
			"callback_id": upd.Callback.CallbackID,
			"recipient":   recipient,
			"sender":      upd.Callback.User,
			"sender_id":   fmt.Sprintf("%d", upd.Callback.User.UserId),
			"message_id":  messageID,
			"message_seq": messageSeq,
		},
	}

	if err := handler.Handle(ctx, req, b); err != nil {
		logger.Error().Err(err).Msg("callback handler failed")
	}

	// Сохраняем состояние после обработки callback
	// Handler может изменить userState через указатель, поэтому сохраняем его после обработки
	if userID != "" && req.UserState != nil {
		stateToSave := *req.UserState
		stateToSave.LastUpdated = time.Now()
		if stateToSave.UserRegistrationData == nil {
			stateToSave.UserRegistrationData = make(map[string]string)
		}
		if err := b.state.SaveUserState(ctx, userID, stateToSave); err != nil {
			logger.Error().Err(err).Msg("failed to save user state")
		}
	}
}
