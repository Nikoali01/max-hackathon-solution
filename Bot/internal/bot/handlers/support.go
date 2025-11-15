package handlers

import (
	"context"
	"strings"

	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/support"
)

type SupportHandler struct {
	service support.Service
	logger  zerolog.Logger
}

func NewSupportHandler(service support.Service, logger zerolog.Logger) *SupportHandler {
	return &SupportHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SupportHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	if strings.TrimSpace(req.Args) == "" {
		return responder.SendText(ctx, req.Recipient(), "Чтобы отправить обращение, напиши /contact <тема>:<сообщение>.\nНапример: /contact Справка:Нужна справка для военкомата.")
	}

	parts := strings.SplitN(req.Args, ":", 2)
	if len(parts) < 2 {
		return responder.SendText(ctx, req.Recipient(), "Пожалуйста, укажи тему и сообщение через двоеточие. Пример: /contact Стипендия:Не пришла стипендия за ноябрь.")
	}

	subject := strings.TrimSpace(parts[0])
	body := strings.TrimSpace(parts[1])
	if subject == "" || body == "" {
		return responder.SendText(ctx, req.Recipient(), "Тема и текст обращения не могут быть пустыми. Попробуй ещё раз.")
	}

	ticket, err := h.service.CreateTicket(ctx, req.UserID(), subject, body)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create ticket")
		return responder.SendText(ctx, req.Recipient(), "Не удалось создать обращение. Попробуй чуть позже или напиши в деканат.")
	}

	response := strings.Builder{}
	response.WriteString("✅ Обращение отправлено в Department of Education.\n")
	response.WriteString("Номер заявки: " + ticket.ID + "\n")
	response.WriteString("Тема: " + ticket.Subject + "\n")
	response.WriteString("Статус: " + ticket.Status + "\n\n")
	response.WriteString("Мы вернёмся с ответом в течение рабочего дня. Я напомню, как только будет обновление.")

	return responder.SendText(ctx, req.Recipient(), response.String())
}
