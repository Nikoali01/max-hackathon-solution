package handlers

import (
	"context"

	"first-max-bot/internal/bot"
)

// OpenDayHandler обрабатывает команду /openday
type OpenDayHandler struct{}

func NewOpenDayHandler() *OpenDayHandler {
	return &OpenDayHandler{}
}

func (h *OpenDayHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := `🎓 День открытых дверей

Ближайший день открытых дверей состоится:
📅 Дата: 15 декабря 2024
🕐 Время: 10:00 - 16:00
📍 Место: Главный корпус университета

В программе:
• Презентация программ обучения
• Экскурсия по кампусу
• Встреча с преподавателями
• Консультации по поступлению

Для записи на день открытых дверей напиши нам: /contact

Мы будем рады видеть тебя! 👋`
	
	return responder.SendText(ctx, req.Recipient(), message)
}

