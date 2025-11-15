package handlers

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/news"
)

// NewsHandler обрабатывает команду /news
type NewsHandler struct {
	newsService news.Service
	logger      zerolog.Logger
}

func NewNewsHandler(newsService news.Service, logger zerolog.Logger) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
		logger:      logger,
	}
}

func (h *NewsHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Получаем последние 3 новости
	latestNews, err := h.newsService.GetLatestNews(ctx, 3)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get latest news")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении новостей")
	}

	if len(latestNews) == 0 {
		return responder.SendText(ctx, req.Recipient(), "📰 Пока нет новостей.")
	}

	// Отправляем каждую новость отдельным сообщением
	for _, n := range latestNews {
		message := fmt.Sprintf("**%s**\n\n%s\n\n_%s, %s_",
			n.Title,
			n.Content,
			n.Author,
			n.CreatedAt.Format("02.01.2006 15:04"),
		)

		if err := responder.SendMarkdown(ctx, req.Recipient(), message); err != nil {
			h.logger.Warn().Err(err).Str("news_id", n.ID).Msg("failed to send news")
		}
	}

	return nil
}
