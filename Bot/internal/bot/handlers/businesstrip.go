package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/businesstrip"
)

// BusinessTripHandler обрабатывает команду /businesstrip для сотрудников
type BusinessTripHandler struct {
	tripService businesstrip.Service
	logger      zerolog.Logger
}

func NewBusinessTripHandler(tripService businesstrip.Service, logger zerolog.Logger) *BusinessTripHandler {
	return &BusinessTripHandler{
		tripService: tripService,
		logger:      logger,
	}
}

func (h *BusinessTripHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userID := req.UserID()
	if userID == "" {
		return responder.SendText(ctx, req.Recipient(), "Не удалось определить пользователя")
	}

	// Получаем командировки пользователя
	trips, err := h.tripService.GetUserTrips(ctx, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get user trips")
	}

	var message strings.Builder
	message.WriteString("✈️ Командировки\n\n")

	if len(trips) > 0 {
		message.WriteString("📋 Твои командировки:\n\n")
		for _, trip := range trips {
			statusEmoji := h.getStatusEmoji(trip.Status)
			message.WriteString(fmt.Sprintf("%s %s\n", statusEmoji, trip.Destination))
			message.WriteString(fmt.Sprintf("   %s - %s\n", trip.StartDate.Format("02.01.2006"), trip.EndDate.Format("02.01.2006")))
			message.WriteString(fmt.Sprintf("   Статус: %s\n\n", trip.Status))
		}
	} else {
		message.WriteString("У тебя пока нет командировок.\n\n")
	}

	message.WriteString("Для оформления новой командировки напиши: /contact\n")
	message.WriteString("Укажи:\n")
	message.WriteString("• Куда (город/страна)\n")
	message.WriteString("• Цель командировки\n")
	message.WriteString("• Даты (начало и конец)")

	return responder.SendText(ctx, req.Recipient(), message.String())
}

func (h *BusinessTripHandler) getStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "approved":
		return "✅"
	case "rejected":
		return "❌"
	case "completed":
		return "✅"
	default:
		return "📄"
	}
}

