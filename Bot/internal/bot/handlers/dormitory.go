package handlers

import (
	"context"

	"first-max-bot/internal/bot"
)

// DormitoryHandler обрабатывает команду /dormitory
type DormitoryHandler struct{}

func NewDormitoryHandler() *DormitoryHandler {
	return &DormitoryHandler{}
}

func (h *DormitoryHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := `🏠 Общежитие

Доступные услуги:
• Оплатить проживание
• Заказать дополнительные услуги
• Оформить пропуск для гостя
• Подать заявку в техподдержку

(Функционал будет добавлен позже)

Для получения помощи по вопросам общежития напиши: /contact`
	
	return responder.SendText(ctx, req.Recipient(), message)
}

