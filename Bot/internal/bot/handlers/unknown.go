package handlers

import (
	"context"

	"first-max-bot/internal/bot"
)

type FallbackHandler struct{}

func NewFallbackHandler() *FallbackHandler {
	return &FallbackHandler{}
}

func (h *FallbackHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	return responder.SendText(ctx, req.Recipient(), "Я пока не знаю такой команды. Попробуй /help, чтобы посмотреть, что я уже умею.")
}
