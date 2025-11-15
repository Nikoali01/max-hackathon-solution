package handlers

import (
	"context"

	"first-max-bot/internal/bot"
)

// AdmissionHandler обрабатывает команды для абитуриентов
type AdmissionHandler struct{}

func NewAdmissionHandler() *AdmissionHandler {
	return &AdmissionHandler{}
}

func (h *AdmissionHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := `📚 Информация о поступлении

Наш университет предлагает широкий спектр образовательных программ.

Для получения подробной информации:
• Ознакомься с программами обучения: /programs
• Запишись на день открытых дверей: /openday
• Задай вопрос: /contact

Мы поможем тебе выбрать подходящую программу! 🎓`
	
	return responder.SendText(ctx, req.Recipient(), message)
}

