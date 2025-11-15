package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/moodle"
	"first-max-bot/internal/services/user"
	"first-max-bot/internal/state"
)

// MoodleHandler обрабатывает команду /moodle
type MoodleHandler struct {
	moodleService moodle.Service
	userService   user.Service
	logger        zerolog.Logger
}

func NewMoodleHandler(moodleService moodle.Service, userService user.Service, logger zerolog.Logger) *MoodleHandler {
	return &MoodleHandler{
		moodleService: moodleService,
		userService:   userService,
		logger:        logger,
	}
}

func (h *MoodleHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userID := req.UserID()

	// Если это callback
	if strings.HasPrefix(req.Args, "moodle:") {
		return h.HandleCallback(ctx, req, responder)
	}

	// Если это текстовый ввод для токена
	if req.UserState != nil && req.UserState.UserRegistrationStep == "moodle_token" {
		return h.HandleTextInput(ctx, req, responder)
	}

	// Проверяем, что пользователь - студент
	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil || u == nil {
		return responder.SendText(ctx, req.Recipient(), "❌ Пользователь не найден")
	}

	if u.Role != user.RoleStudent {
		return responder.SendText(ctx, req.Recipient(), "❌ Эта команда доступна только студентам.")
	}

	// Проверяем, есть ли токен
	if u.MoodleToken == "" {
		// Если пользователь в процессе ввода токена
		if req.UserState != nil && req.UserState.UserRegistrationStep == "moodle_token" {
			return h.handleTokenInput(ctx, req, responder)
		}

		// Предлагаем добавить токен
		if req.UserState == nil {
			req.UserState = &state.UserState{}
		}
		req.UserState.UserRegistrationStep = "moodle_token"

		message := "🔗 **Интеграция с Moodle**\n\n"
		message += "Для работы с Moodle необходимо добавить токен доступа.\n\n"
		message += "Введи свой токен Moodle:"

		return responder.SendMarkdown(ctx, req.Recipient(), message)
	}

	// Если токен есть, получаем информацию о пользователе
	siteInfo, err := h.moodleService.GetSiteInfo(ctx, u.MoodleToken)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("failed to get moodle site info")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при подключении к Moodle. Проверь токен или попробуй позже.")
	}

	// Формируем сообщение с информацией о пользователе
	var message strings.Builder
	message.WriteString("🔗 **Moodle**\n\n")
	message.WriteString(fmt.Sprintf("**Сайт:** %s\n", siteInfo.Sitename))
	message.WriteString(fmt.Sprintf("**Пользователь:** %s\n", siteInfo.Fullname))
	message.WriteString(fmt.Sprintf("**Логин:** %s\n", siteInfo.Username))
	message.WriteString(fmt.Sprintf("**Версия:** %s\n\n", siteInfo.Release))

	// Показываем доступные функции (первые 5)
	if len(siteInfo.Functions) > 0 {
		message.WriteString("**Доступные функции:**\n")
		maxFuncs := 5
		if len(siteInfo.Functions) < maxFuncs {
			maxFuncs = len(siteInfo.Functions)
		}
		for i := 0; i < maxFuncs; i++ {
			message.WriteString(fmt.Sprintf("• %s\n", siteInfo.Functions[i].Name))
		}
		if len(siteInfo.Functions) > maxFuncs {
			message.WriteString(fmt.Sprintf("... и ещё %d\n", len(siteInfo.Functions)-maxFuncs))
		}
		message.WriteString("\n")
	}

	// Создаем клавиатуру с возможностями
	keyboard := responder.NewKeyboardBuilder()
	row := keyboard.AddRow()
	row.AddCallback("🔄 Обновить информацию", schemes.POSITIVE, "moodle:refresh")
	row2 := keyboard.AddRow()
	row2.AddCallback("🔑 Изменить токен", schemes.POSITIVE, "moodle:change_token")
	row3 := keyboard.AddRow()
	row3.AddCallback("📚 Мои курсы", schemes.POSITIVE, "moodle:courses")

	return responder.SendMarkdownWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
}

func (h *MoodleHandler) HandleTextInput(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	if req.UserState == nil || req.UserState.UserRegistrationStep != "moodle_token" {
		return nil
	}

	return h.handleTokenInput(ctx, req, responder)
}

func (h *MoodleHandler) handleTokenInput(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userID := req.UserID()
	token := strings.TrimSpace(req.Args)

	if token == "" {
		return responder.SendText(ctx, req.Recipient(), "❌ Токен не может быть пустым. Попробуй снова.")
	}

	// Проверяем токен, делая запрос к Moodle
	siteInfo, err := h.moodleService.GetSiteInfo(ctx, token)
	if err != nil {
		h.logger.Warn().Err(err).Str("user_id", userID).Msg("invalid moodle token")
		return responder.SendText(ctx, req.Recipient(), "❌ Неверный токен. Проверь правильность токена и попробуй снова.")
	}

	// Сохраняем токен
	if err := h.userService.SetMoodleToken(ctx, userID, token); err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("failed to save moodle token")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при сохранении токена.")
	}

	// Очищаем состояние
	req.UserState.UserRegistrationStep = ""

	message := fmt.Sprintf("✅ Токен успешно привязан!\n\n")
	message += fmt.Sprintf("**Пользователь:** %s\n", siteInfo.Fullname)
	message += fmt.Sprintf("**Сайт:** %s\n\n", siteInfo.Sitename)
	message += "Теперь ты можешь использовать все возможности Moodle."

	return responder.SendMarkdown(ctx, req.Recipient(), message)
}

func (h *MoodleHandler) HandleCallback(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	payload := req.Args
	callbackID := ""

	// Получаем callback ID из Metadata (он там сохраняется в handleCallback)
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok {
			callbackID = cid
			responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
		}
	}

	userID := req.UserID()
	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil || u == nil || u.MoodleToken == "" {
		return responder.SendText(ctx, req.Recipient(), "❌ Токен Moodle не найден. Используй /moodle для привязки.")
	}

	if strings.HasPrefix(payload, "moodle:refresh") {
		// Обновляем информацию
		siteInfo, err := h.moodleService.GetSiteInfo(ctx, u.MoodleToken)
		if err != nil {
			h.logger.Error().Err(err).Str("user_id", userID).Msg("failed to refresh moodle info")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при обновлении информации.")
		}

		message := fmt.Sprintf("✅ Информация обновлена!\n\n")
		message += fmt.Sprintf("**Пользователь:** %s\n", siteInfo.Fullname)
		message += fmt.Sprintf("**Логин:** %s\n", siteInfo.Username)
		message += fmt.Sprintf("**Сайт:** %s\n", siteInfo.Sitename)
		message += fmt.Sprintf("**Версия:** %s", siteInfo.Release)

		return responder.SendMarkdown(ctx, req.Recipient(), message)
	}

	if strings.HasPrefix(payload, "moodle:change_token") {
		// Начинаем процесс смены токена
		if req.UserState == nil {
			req.UserState = &state.UserState{}
		}
		req.UserState.UserRegistrationStep = "moodle_token"

		message := "🔑 **Изменение токена Moodle**\n\n"
		message += "Введи новый токен:"

		return responder.SendMarkdown(ctx, req.Recipient(), message)
	}

	if strings.HasPrefix(payload, "moodle:courses") {
		// Получаем информацию о пользователе для получения userID
		siteInfo, err := h.moodleService.GetSiteInfo(ctx, u.MoodleToken)
		if err != nil {
			h.logger.Error().Err(err).Str("user_id", userID).Msg("failed to get site info for courses")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении информации о пользователе.")
		}

		// Получаем курсы пользователя
		courses, err := h.moodleService.GetUserCourses(ctx, u.MoodleToken, siteInfo.UserID)
		if err != nil {
			h.logger.Error().Err(err).Str("user_id", userID).Int("moodle_user_id", siteInfo.UserID).Msg("failed to get user courses")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении курсов.")
		}

		if len(courses) == 0 {
			return responder.SendMarkdown(ctx, req.Recipient(), "📚 У тебя пока нет курсов в Moodle.")
		}

		// Отправляем каждый курс отдельным сообщением
		for i, course := range courses {
			if i >= 10 { // Ограничиваем до 10 курсов
				break
			}

			var message strings.Builder
			message.WriteString(fmt.Sprintf("📚 **%s**\n\n", course.Fullname))
			
			if course.Summary != "" {
				// Упрощаем HTML описание - убираем теги для краткости
				summary := course.Summary
				// Убираем HTML теги (простая замена)
				summary = strings.ReplaceAll(summary, "<br />", "\n")
				summary = strings.ReplaceAll(summary, "<br>", "\n")
				summary = strings.ReplaceAll(summary, "</h3>", "\n")
				summary = strings.ReplaceAll(summary, "</h5>", "\n")
				summary = strings.ReplaceAll(summary, "<h3>", "")
				summary = strings.ReplaceAll(summary, "<h5>", "")
				summary = strings.ReplaceAll(summary, "<strong>", "**")
				summary = strings.ReplaceAll(summary, "</strong>", "**")
				summary = strings.ReplaceAll(summary, "&nbsp;", " ")
				// Убираем все остальные HTML теги (простой подход)
				for strings.Contains(summary, "<") && strings.Contains(summary, ">") {
					start := strings.Index(summary, "<")
					end := strings.Index(summary[start:], ">")
					if end != -1 {
						summary = summary[:start] + summary[start+end+1:]
					} else {
						break
					}
				}
				
				// Очищаем от лишних пробелов и переносов
				summary = strings.TrimSpace(summary)
				summary = strings.ReplaceAll(summary, "\n\n\n", "\n\n")
				
				// Берем первые 300 символов описания
				if len(summary) > 300 {
					summary = summary[:300] + "..."
				}
				if summary != "" {
					message.WriteString(fmt.Sprintf("%s\n\n", summary))
				}
			}

			// Даты
			if course.StartDate > 0 {
				startDate := time.Unix(course.StartDate, 0)
				message.WriteString(fmt.Sprintf("📅 Начало: %s\n", startDate.Format("02.01.2006")))
			}
			if course.EndDate > 0 {
				endDate := time.Unix(course.EndDate, 0)
				message.WriteString(fmt.Sprintf("📅 Окончание: %s\n", endDate.Format("02.01.2006")))
			}

			// Прогресс
			if course.Progress != nil {
				message.WriteString(fmt.Sprintf("📊 Прогресс: %d%%\n", *course.Progress))
			}

			// Статус
			if course.Completed {
				message.WriteString("✅ Завершен\n")
			} else {
				message.WriteString("⏳ В процессе\n")
			}

			// Последний доступ
			if course.LastAccess > 0 {
				lastAccess := time.Unix(course.LastAccess, 0)
				message.WriteString(fmt.Sprintf("🕐 Последний доступ: %s", lastAccess.Format("02.01.2006 15:04")))
			}

			if err := responder.SendMarkdown(ctx, req.Recipient(), message.String()); err != nil {
				h.logger.Warn().Err(err).Int("course_id", course.ID).Msg("failed to send course info")
			}
		}

		if len(courses) > 10 {
			message := fmt.Sprintf("\n... и ещё %d курсов", len(courses)-10)
			responder.SendMarkdown(ctx, req.Recipient(), message)
		}

		return nil
	}

	return nil
}

