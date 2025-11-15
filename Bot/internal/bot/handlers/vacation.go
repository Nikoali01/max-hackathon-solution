package handlers

import (
	"context"

	"first-max-bot/internal/bot"
)

// VacationHandler обрабатывает команду /vacation для сотрудников
type VacationHandler struct{}

func NewVacationHandler() *VacationHandler {
	return &VacationHandler{}
}

func (h *VacationHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := `🏖️ Отпуска

Здесь ты можешь:
• Оформить заявку на отпуск
• Согласовать отпуск
• Посмотреть график отпусков

(Функционал будет добавлен позже)

Для оформления отпуска напиши: /contact`
	
	return responder.SendText(ctx, req.Recipient(), message)
}

