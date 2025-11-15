package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/schedule"
)

type ScheduleHandler struct {
	service schedule.Service
	logger  zerolog.Logger
}

func NewScheduleHandler(service schedule.Service, logger zerolog.Logger) *ScheduleHandler {
	return &ScheduleHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ScheduleHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userID := req.UserID()
	items, err := h.service.GetSchedule(ctx, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get schedule")
		return responder.SendText(ctx, req.Recipient(), "Не удалось получить расписание. Попробуйте позже.")
	}

	if len(items) == 0 {
		return responder.SendText(ctx, req.Recipient(), "Расписание на сегодня пустое. Используйте /contact, если нужен совет.")
	}

	var b strings.Builder
	b.WriteString("📅 Ваше расписание на сегодня:\n\n")
	for _, item := range items {
		b.WriteString(fmt.Sprintf(
			"• %s — %s\n  %s, %s\n  %s\n\n",
			item.Time.Format("15:04"),
			item.Discipline,
			item.Instructor,
			item.Location,
			item.Description,
		))
	}

	return responder.SendText(ctx, req.Recipient(), b.String())
}
