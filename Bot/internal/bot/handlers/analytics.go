package handlers

import (
	"context"

	"first-max-bot/internal/bot"
)

// AnalyticsHandler обрабатывает команду /analytics для руководителей
type AnalyticsHandler struct{}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{}
}

func (h *AnalyticsHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := `📈 Аналитика

Доступные показатели:
• Научные показатели вуза
• Академические показатели
• Статистика по программам
• Показатели вовлеченности студентов

(Функционал будет добавлен позже)

Для получения новостей используй: /news`
	
	return responder.SendText(ctx, req.Recipient(), message)
}

