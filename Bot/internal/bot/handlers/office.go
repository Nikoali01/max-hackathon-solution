package handlers

import (
	"context"

	"first-max-bot/internal/bot"
)

// OfficeHandler обрабатывает команду /office для сотрудников
type OfficeHandler struct{}

func NewOfficeHandler() *OfficeHandler {
	return &OfficeHandler{}
}

func (h *OfficeHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := `🏢 Офис

Доступные услуги:
• Заказать справку с места работы
• Оформить гостевой пропуск в офис
• Получить доступ к офисным помещениям

(Функционал будет добавлен позже)

Для получения справки или оформления пропуска напиши: /contact`
	
	return responder.SendText(ctx, req.Recipient(), message)
}

