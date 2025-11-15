package handlers

import (
	"context"
	"strings"

	"first-max-bot/internal/bot"
)

type HelpHandler struct {
	Lines []string
}

func NewHelpHandler() *HelpHandler {
	return &HelpHandler{
		Lines: []string{
			"🚀 Полный список команд:",
			"/schedule — увидеть расписание и полезные подсказки.",
			"/contact <тема>:<сообщение> — отправить обращение в Department of Education.",
			"/help — короткая справка о возможностях помощника.",
		},
	}
}

func (h *HelpHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := strings.Join(h.Lines, "\n")
	return responder.SendText(ctx, req.Recipient(), message)
}
