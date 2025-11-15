package handlers

import (
	"context"

	"first-max-bot/internal/bot"
)

// DashboardHandler обрабатывает команду /dashboard для руководителей
type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := `📊 Дашборд

Доступная информация:
• Количество студентов и сотрудников в кампусе
• Интеграция с системой контроля доступа
• Научные и академические показатели вуза

(Функционал будет добавлен позже)

Для получения подробной аналитики используй: /analytics`
	
	return responder.SendText(ctx, req.Recipient(), message)
}

