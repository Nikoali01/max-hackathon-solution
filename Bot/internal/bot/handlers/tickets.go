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
	"first-max-bot/internal/services/user"
	"first-max-bot/internal/state"
)

// TicketsHandler обрабатывает команду /tickets для руководителей
type TicketsHandler struct {
	supportService support.Service
	userService    user.Service
	logger         zerolog.Logger
}

func NewTicketsHandler(supportService support.Service, userService user.Service, logger zerolog.Logger) *TicketsHandler {
	return &TicketsHandler{
		supportService: supportService,
		userService:    userService,
		logger:         logger,
	}
}

func (h *TicketsHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Проверяем, что пользователь - руководитель
	userID := req.UserID()
	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil || u == nil || u.Role != user.RoleManager {
		return responder.SendText(ctx, req.Recipient(), "❌ Эта команда доступна только руководителям.")
	}

	// Если это текстовый ввод для ответа на обращение
	userState := req.UserState
	if userState != nil && userState.UserRegistrationStep == "ticket_reply" {
		return h.HandleTextInput(ctx, req, responder)
	}

	// Если это callback для ответа на обращение
	if strings.HasPrefix(req.Args, "ticket:") {
		return h.handleTicketCallback(ctx, req, responder)
	}

	// Получаем все обращения
	tickets, err := h.supportService.GetAllTickets(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get tickets")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении обращений")
	}

	// Фильтруем нерешенные обращения (не закрытые)
	var pendingTickets []support.Ticket
	for _, ticket := range tickets {
		if ticket.Status != "closed" && ticket.Status != "resolved" {
			pendingTickets = append(pendingTickets, ticket)
		}
	}

	var message strings.Builder
	message.WriteString("📋 Обращения\n\n")

	if len(pendingTickets) == 0 {
		message.WriteString("✅ Нет нерешенных обращений.")
		return responder.SendText(ctx, req.Recipient(), message.String())
	}

	message.WriteString(fmt.Sprintf("Нерешенных обращений: %d\n\n", len(pendingTickets)))

	keyboard := responder.NewKeyboardBuilder()

	// Показываем первые 5 обращений
	for i, ticket := range pendingTickets {
		if i >= 5 {
			break
		}
		
		row := keyboard.AddRow()
		subject := ticket.Subject
		if len(subject) > 30 {
			subject = subject[:30] + "..."
		}
		row.AddCallback(fmt.Sprintf("📄 %s", subject), schemes.DEFAULT, fmt.Sprintf("ticket:view:%s", ticket.ID))
	}

	return responder.SendTextWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
}

func (h *TicketsHandler) handleTicketCallback(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	payload := req.Args

	callbackID := ""
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok && cid != "" {
			callbackID = cid
			responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
		}
	}

	if strings.HasPrefix(payload, "ticket:view:") {
		ticketID := strings.TrimPrefix(payload, "ticket:view:")
		
		ticket, err := h.supportService.GetTicket(ctx, ticketID)
		if err != nil || ticket == nil {
			return responder.SendText(ctx, req.Recipient(), "❌ Обращение не найдено")
		}

		var message strings.Builder
		message.WriteString(fmt.Sprintf("📄 Обращение #%s\n\n", ticket.ID))
		message.WriteString(fmt.Sprintf("Тема: %s\n", ticket.Subject))
		message.WriteString(fmt.Sprintf("От: %s\n", ticket.UserID))
		message.WriteString(fmt.Sprintf("Статус: %s\n", ticket.Status))
		message.WriteString(fmt.Sprintf("Создано: %s\n\n", ticket.CreatedAt.Format("02.01.2006 15:04")))
		message.WriteString(fmt.Sprintf("Сообщение:\n%s\n\n", ticket.Message))

		if ticket.Response != "" {
			message.WriteString(fmt.Sprintf("📤 Ответ руководителя:\n%s\n\n", ticket.Response))
		} else {
			message.WriteString("Ответ ещё не дан.\n\n")
		}

		if ticket.UserReply != "" {
			message.WriteString(fmt.Sprintf("📥 Ответы пользователя:\n%s\n\n", ticket.UserReply))
		}

		keyboard := responder.NewKeyboardBuilder()
		// Показываем кнопку "Ответить" если еще нет ответа или если пользователь ответил на ответ
		if ticket.Response == "" || (ticket.Response != "" && ticket.UserReply != "") {
			row := keyboard.AddRow()
			row.AddCallback("✍️ Ответить", schemes.POSITIVE, fmt.Sprintf("ticket:reply:%s", ticket.ID))
		}
		row := keyboard.AddRow()
		row.AddCallback("✅ Закрыть", schemes.POSITIVE, fmt.Sprintf("ticket:close:%s", ticket.ID))

		return responder.SendTextWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
	}

	if strings.HasPrefix(payload, "ticket:reply:") {
		ticketID := strings.TrimPrefix(payload, "ticket:reply:")
		
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
		req.UserState.UserRegistrationStep = "ticket_reply"

		message := "✍️ Напиши ответ на обращение:"
		return responder.SendText(ctx, req.Recipient(), message)
	}

	if strings.HasPrefix(payload, "ticket:close:") {
		ticketID := strings.TrimPrefix(payload, "ticket:close:")
		
		// Получаем тикет перед закрытием, чтобы отправить уведомление пользователю
		ticket, err := h.supportService.GetTicket(ctx, ticketID)
		if err != nil || ticket == nil {
			return responder.SendText(ctx, req.Recipient(), "❌ Обращение не найдено")
		}
		
		err = h.supportService.UpdateTicketStatus(ctx, ticketID, "closed")
		if err != nil {
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при закрытии обращения")
		}

		// Отправляем уведомление пользователю о закрытии тикета
		userIDInt, err := strconv.ParseInt(ticket.UserID, 10, 64)
		if err == nil {
			userRecipient := schemes.Recipient{
				UserId:   userIDInt,
				ChatType: schemes.DIALOG,
			}
			
			notification := fmt.Sprintf("🔒 Твоё обращение #%s закрыто\n\n", ticketID)
			notification += fmt.Sprintf("Тема: %s\n\n", ticket.Subject)
			notification += "Обращение закрыто администратором. Если у тебя есть дополнительные вопросы, создай новое обращение через /contact"
			
			// Отправляем уведомление пользователю
			if err := responder.SendText(ctx, userRecipient, notification); err != nil {
				h.logger.Warn().Err(err).Str("user_id", ticket.UserID).Msg("failed to send closure notification to user")
			} else {
				h.logger.Info().Str("ticket_id", ticketID).Str("user_id", ticket.UserID).Msg("user notified about ticket closure")
			}
		}

		return responder.SendText(ctx, req.Recipient(), "✅ Обращение закрыто. Пользователь получит уведомление.")
	}

	return nil
}

// HandleTextInput обрабатывает текстовый ввод для ответа на обращение
func (h *TicketsHandler) HandleTextInput(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userState := req.UserState
	
	// Обработка текстового ответа на обращение
	if userState != nil && userState.UserRegistrationStep == "ticket_reply" {
		ticketID := userState.UserRegistrationData["replying_to_ticket"]
		if ticketID == "" {
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка: не найден ID обращения")
		}

		responseText := strings.TrimSpace(req.Args)
		if responseText == "" {
			return responder.SendText(ctx, req.Recipient(), "❌ Ответ не может быть пустым")
		}

		// Получаем ID текущего администратора
		adminUserID := req.UserID()
		err := h.supportService.AddResponse(ctx, ticketID, responseText, adminUserID)
		if err != nil {
			h.logger.Error().Err(err).Msg("failed to add response")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при сохранении ответа")
		}

		// Получаем тикет для отправки уведомления пользователю
		ticket, err := h.supportService.GetTicket(ctx, ticketID)
		if err == nil && ticket != nil {
			// Отправляем уведомление пользователю
			userIDInt, err := strconv.ParseInt(ticket.UserID, 10, 64)
			if err == nil {
				userRecipient := schemes.Recipient{
					UserId:   userIDInt,
					ChatType: schemes.DIALOG,
				}
				
				notification := fmt.Sprintf("📬 Новый ответ на твоё обращение #%s\n\n", ticketID)
				notification += fmt.Sprintf("Тема: %s\n\n", ticket.Subject)
				notification += fmt.Sprintf("Ответ:\n%s\n\n", responseText)
				notification += "Используй /mytickets чтобы посмотреть все свои обращения и ответить."
				
				// Отправляем уведомление пользователю
				if err := responder.SendText(ctx, userRecipient, notification); err != nil {
					h.logger.Warn().Err(err).Str("user_id", ticket.UserID).Msg("failed to send notification to user")
				} else {
					h.logger.Info().Str("ticket_id", ticketID).Str("user_id", ticket.UserID).Msg("user notified about response")
				}
			}
		}

		// Очищаем состояние
		userState.UserRegistrationStep = ""
		delete(userState.UserRegistrationData, "replying_to_ticket")

		message := fmt.Sprintf("✅ Ответ на обращение #%s сохранён!\n\n", ticketID)
		message += fmt.Sprintf("Ответ:\n%s\n\n", responseText)
		message += "Пользователь получит уведомление. Тикет остаётся открытым до явного закрытия."

		return responder.SendText(ctx, req.Recipient(), message)
	}

	return nil
}

