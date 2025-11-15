package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/news"
	"first-max-bot/internal/services/user"
	"first-max-bot/internal/state"
)

// SendNewsHandler обрабатывает команду для отправки новостей администратором
type SendNewsHandler struct {
	newsService news.Service
	userService user.Service
	logger      zerolog.Logger
}

func NewSendNewsHandler(newsService news.Service, userService user.Service, logger zerolog.Logger) *SendNewsHandler {
	return &SendNewsHandler{
		newsService: newsService,
		userService: userService,
		logger:      logger,
	}
}

func (h *SendNewsHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userID := req.UserID()

	// Проверяем, что пользователь - администратор (manager или head)
	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil || u == nil {
		return responder.SendText(ctx, req.Recipient(), "❌ Пользователь не найден")
	}

	if u.Role != user.RoleManager {
		return responder.SendText(ctx, req.Recipient(), "❌ Эта команда доступна только администраторам.")
	}

	// Проверяем состояние - если пользователь уже в процессе отправки новости
	if req.UserState != nil && req.UserState.UserRegistrationStep == "send_news" {
		return h.handleNewsContent(ctx, req, responder)
	}

	// Начинаем процесс отправки новости
	if req.UserState == nil {
		req.UserState = &state.UserState{}
	}
	req.UserState.UserRegistrationStep = "send_news"
	if req.UserState.UserRegistrationData == nil {
		req.UserState.UserRegistrationData = make(map[string]string)
	}

	message := "📰 **Отправка новости**\n\n"
	message += "Введи заголовок новости:"

	return responder.SendMarkdown(ctx, req.Recipient(), message)
}

func (h *SendNewsHandler) HandleTextInput(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	if req.UserState == nil || req.UserState.UserRegistrationStep != "send_news" {
		return nil
	}

	text := strings.TrimSpace(req.Args)
	if text == "" {
		return responder.SendText(ctx, req.Recipient(), "❌ Текст не может быть пустым. Попробуй снова.")
	}

	// Определяем, на каком шаге мы находимся
	if req.UserState.UserRegistrationData == nil {
		req.UserState.UserRegistrationData = make(map[string]string)
	}

	// Если заголовок еще не введен
	if req.UserState.UserRegistrationData["title"] == "" {
		req.UserState.UserRegistrationData["title"] = text
		message := "✅ Заголовок сохранён.\n\n"
		message += "Теперь введи текст новости (в markdown формате):"
		return responder.SendText(ctx, req.Recipient(), message)
	}

	// Если заголовок уже есть, значит это текст новости
	title := req.UserState.UserRegistrationData["title"]
	content := text

	// Получаем информацию об авторе
	userID := req.UserID()
	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil || u == nil {
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка: пользователь не найден")
	}

	authorName := fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	if authorName == " " {
		authorName = "Администратор"
	}

	// Создаем новость
	newsItem, err := h.newsService.CreateNews(ctx, title, content, userID, authorName)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create news")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при создании новости")
	}

	// Формируем сообщение для отправки
	newsMessage := fmt.Sprintf("📰 **%s**\n\n%s\n\n_%s, %s_",
		newsItem.Title,
		newsItem.Content,
		newsItem.Author,
		newsItem.CreatedAt.Format("02.01.2006 15:04"),
	)

	// Получаем всех пользователей
	allUsers, err := h.userService.GetAllUsers(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get all users")
		// Очищаем состояние
		req.UserState.UserRegistrationStep = ""
		req.UserState.UserRegistrationData = nil
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении списка пользователей")
	}

	// Отправляем новость всем пользователям (кроме отправителя)
	sentCount := 0
	failedCount := 0
	for _, u := range allUsers {
		// Пропускаем отправителя
		if u.UserID == userID {
			continue
		}

		userIDInt, err := strconv.ParseInt(u.UserID, 10, 64)
		if err != nil {
			h.logger.Warn().Str("user_id", u.UserID).Err(err).Msg("failed to parse user ID")
			failedCount++
			continue
		}

		recipient := schemes.Recipient{
			UserId:   userIDInt,
			ChatType: schemes.DIALOG,
		}

		if err := responder.SendMarkdown(ctx, recipient, newsMessage); err != nil {
			h.logger.Warn().Err(err).Str("user_id", u.UserID).Msg("failed to send news to user")
			failedCount++
		} else {
			sentCount++
		}
	}

	// Очищаем состояние
	req.UserState.UserRegistrationStep = ""
	req.UserState.UserRegistrationData = nil

	message := fmt.Sprintf("✅ Новость создана и отправлена!\n\n")
	message += fmt.Sprintf("**%s**\n\n%s\n\n", title, content)
	message += fmt.Sprintf("Отправлено: %d пользователей\n", sentCount)
	if failedCount > 0 {
		message += fmt.Sprintf("Ошибок: %d\n", failedCount)
	}

	return responder.SendMarkdown(ctx, req.Recipient(), message)
}

func (h *SendNewsHandler) handleNewsContent(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	return h.HandleTextInput(ctx, req, responder)
}
