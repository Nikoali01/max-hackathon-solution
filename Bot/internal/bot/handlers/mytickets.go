package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/support"
	"first-max-bot/internal/state"
)

// MyTicketsHandler обрабатывает команду /mytickets для пользователей
type MyTicketsHandler struct {
	supportService support.Service
	logger         zerolog.Logger
}

func NewMyTicketsHandler(supportService support.Service, logger zerolog.Logger) *MyTicketsHandler {
	return &MyTicketsHandler{
		supportService: supportService,
		logger:         logger,
	}
}

func (h *MyTicketsHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Если это текстовый ввод для ответа на обращение
	userState := req.UserState
	if userState != nil && userState.UserRegistrationStep == "ticket_user_reply" {
		return h.HandleTextInput(ctx, req, responder)
	}

	// Если это callback для просмотра или ответа на обращение
	if strings.HasPrefix(req.Args, "myticket:") {
		return h.handleTicketCallback(ctx, req, responder)
	}

	userID := req.UserID()
	if userID == "" {
		return responder.SendText(ctx, req.Recipient(), "Не удалось определить пользователя")
	}

	// Получаем обращения пользователя
	tickets, err := h.supportService.GetUserTickets(ctx, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get user tickets")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении обращений")
	}

	if len(tickets) == 0 {
		return responder.SendText(ctx, req.Recipient(), "📋 У тебя пока нет обращений.\n\nИспользуй /contact чтобы создать обращение.")
	}

	var message strings.Builder
	message.WriteString("📋 Твои обращения\n\n")

	keyboard := responder.NewKeyboardBuilder()

	// Показываем все обращения пользователя
	for i, ticket := range tickets {
		if i >= 10 {
			break
		}

		statusEmoji := h.getStatusEmoji(ticket.Status)
		row := keyboard.AddRow()
		subject := ticket.Subject
		if len(subject) > 25 {
			subject = subject[:25] + "..."
		}
		row.AddCallback(fmt.Sprintf("%s %s", statusEmoji, subject), schemes.DEFAULT, fmt.Sprintf("myticket:view:%s", ticket.ID))
	}

	return responder.SendTextWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
}

func (h *MyTicketsHandler) handleTicketCallback(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	payload := req.Args

	callbackID := ""
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok && cid != "" {
			callbackID = cid
			responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
		}
	}

	if strings.HasPrefix(payload, "myticket:view:") {
		ticketID := strings.TrimPrefix(payload, "myticket:view:")

		// Проверяем, что тикет принадлежит текущему пользователю
		userID := req.UserID()
		ticket, err := h.supportService.GetTicket(ctx, ticketID)
		if err != nil || ticket == nil {
			h.logger.Warn().Str("ticket_id", ticketID).Str("user_id", userID).Msg("ticket not found or access denied")
			return responder.SendText(ctx, req.Recipient(), "❌ Обращение не найдено")
		}

		// Проверяем, что тикет принадлежит пользователю
		if ticket.UserID != userID {
			h.logger.Warn().Str("ticket_id", ticketID).Str("user_id", userID).Str("ticket_user_id", ticket.UserID).Msg("user trying to access someone else's ticket")
			return responder.SendText(ctx, req.Recipient(), "❌ У тебя нет доступа к этому обращению")
		}

		var message strings.Builder
		message.WriteString(fmt.Sprintf("📄 Обращение #%s\n\n", ticket.ID))
		message.WriteString(fmt.Sprintf("Тема: %s\n", ticket.Subject))
		message.WriteString(fmt.Sprintf("Статус: %s\n", h.getStatusLabel(ticket.Status)))
		message.WriteString(fmt.Sprintf("Создано: %s\n\n", ticket.CreatedAt.Format("02.01.2006 15:04")))
		message.WriteString(fmt.Sprintf("Твоё сообщение:\n%s\n\n", ticket.Message))

		if ticket.Response != "" {
			message.WriteString(fmt.Sprintf("📤 Ответ:\n%s\n\n", ticket.Response))
		} else {
			message.WriteString("⏳ Ожидаем ответа...\n\n")
		}

		if ticket.UserReply != "" {
			message.WriteString(fmt.Sprintf("📥 Твои ответы:\n%s\n\n", ticket.UserReply))
		}

		keyboard := responder.NewKeyboardBuilder()
		// Если есть ответ руководителя и тикет не закрыт, показываем кнопку "Ответить"
		// Пользователь может отвечать несколько раз, пока тикет не закрыт
		if ticket.Response != "" && ticket.Status != "closed" && ticket.Status != "resolved" {
			row := keyboard.AddRow()
			if ticket.UserReply == "" {
				row.AddCallback("✍️ Ответить на ответ", schemes.POSITIVE, fmt.Sprintf("myticket:reply:%s", ticket.ID))
			} else {
				row.AddCallback("✍️ Ответить снова", schemes.POSITIVE, fmt.Sprintf("myticket:reply:%s", ticket.ID))
			}
		}
		
		// Всегда показываем кнопку для просмотра, даже если тикет закрыт
		// (callback уже обработан выше, это просто для ясности)

		return responder.SendTextWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
	}

	if strings.HasPrefix(payload, "myticket:reply:") {
		ticketID := strings.TrimPrefix(payload, "myticket:reply:")

		// Сохраняем ticketID в состоянии для ответа
		// Важно: изменяем req.UserState напрямую, чтобы изменения сохранились
		if req.UserState == nil {
			req.UserState = &state.UserState{
				UserRegistrationData: make(map[string]string),
			}
		}
		if req.UserState.UserRegistrationData == nil {
			req.UserState.UserRegistrationData = make(map[string]string)
		}
		req.UserState.UserRegistrationData["replying_to_ticket"] = ticketID
		req.UserState.UserRegistrationStep = "ticket_user_reply"

		message := "✍️ Напиши свой ответ на обращение:"
		return responder.SendText(ctx, req.Recipient(), message)
	}

	return nil
}

func (h *MyTicketsHandler) HandleTextInput(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userState := req.UserState

	if userState != nil && userState.UserRegistrationStep == "ticket_user_reply" {
		ticketID := userState.UserRegistrationData["replying_to_ticket"]
		if ticketID == "" {
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка: не найден ID обращения")
		}

		replyText := strings.TrimSpace(req.Args)
		if replyText == "" {
			return responder.SendText(ctx, req.Recipient(), "❌ Ответ не может быть пустым")
		}

		err := h.supportService.AddUserReply(ctx, ticketID, replyText)
		if err != nil {
			h.logger.Error().Err(err).Msg("failed to add user reply")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при сохранении ответа")
		}

		// Получаем тикет для отправки уведомления администратору, который отвечал
		ticket, err := h.supportService.GetTicket(ctx, ticketID)
		if err == nil && ticket != nil && ticket.ResponseBy != "" {
			// Отправляем уведомление администратору, который отвечал
			adminUserIDInt, err := strconv.ParseInt(ticket.ResponseBy, 10, 64)
			if err == nil {
				adminRecipient := schemes.Recipient{
					UserId:   adminUserIDInt,
					ChatType: schemes.DIALOG,
				}

				notification := fmt.Sprintf("📬 Новый ответ на обращение #%s\n\n", ticketID)
				notification += fmt.Sprintf("Тема: %s\n", ticket.Subject)
				notification += fmt.Sprintf("От пользователя: %s\n\n", ticket.UserID)
				notification += fmt.Sprintf("Ответ:\n%s\n\n", replyText)
				notification += "Используй /tickets чтобы посмотреть обращение и ответить."

				// Отправляем уведомление администратору
				if err := responder.SendText(ctx, adminRecipient, notification); err != nil {
					h.logger.Warn().Err(err).Str("admin_user_id", ticket.ResponseBy).Msg("failed to send notification to admin")
				} else {
					h.logger.Info().Str("ticket_id", ticketID).Str("admin_user_id", ticket.ResponseBy).Msg("admin notified about user reply")
				}
			}
		}

		// Очищаем состояние
		userState.UserRegistrationStep = ""
		delete(userState.UserRegistrationData, "replying_to_ticket")

		message := fmt.Sprintf("✅ Твой ответ на обращение #%s сохранён!\n\n", ticketID)
		message += fmt.Sprintf("Ответ:\n%s\n\n", replyText)
		message += "Руководитель получит уведомление о твоём ответе."

		return responder.SendText(ctx, req.Recipient(), message)
	}

	return nil
}

func (h *MyTicketsHandler) getStatusEmoji(status string) string {
	switch status {
	case "received":
		return "📥"
	case "in_progress":
		return "⏳"
	case "answered":
		return "✅"
	case "resolved":
		return "✅"
	case "closed":
		return "🔒"
	default:
		return "📄"
	}
}

func (h *MyTicketsHandler) getStatusLabel(status string) string {
	switch status {
	case "received":
		return "Получено"
	case "in_progress":
		return "В работе"
	case "answered":
		return "Есть ответ"
	case "resolved":
		return "Решено"
	case "closed":
		return "Закрыто"
	default:
		return status
	}
}

