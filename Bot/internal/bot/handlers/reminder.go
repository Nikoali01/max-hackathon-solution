package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/reminder"
	"first-max-bot/internal/state"
)

// ReminderHandler обрабатывает команду /reminder
type ReminderHandler struct {
	reminderService reminder.Service
	logger          zerolog.Logger
}

func NewReminderHandler(reminderService reminder.Service, logger zerolog.Logger) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
		logger:          logger,
	}
}

func (h *ReminderHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Если это callback
	if strings.HasPrefix(req.Args, "reminder:") {
		return h.HandleCallback(ctx, req, responder)
	}

	// Если пользователь в процессе создания напоминания
	if req.UserState != nil && req.UserState.UserRegistrationStep == "reminder_create" {
		return h.HandleTextInput(ctx, req, responder)
	}

	// Показываем меню напоминаний
	return h.showReminderMenu(ctx, req, responder)
}

func (h *ReminderHandler) showReminderMenu(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userID := req.UserID()

	// Получаем активные напоминания пользователя
	reminders, err := h.reminderService.GetActiveReminders(ctx, userID)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("failed to get reminders")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении напоминаний")
	}

	message := "⏰ **Напоминания**\n\n"
	if len(reminders) == 0 {
		message += "У тебя пока нет активных напоминаний.\n\n"
	} else {
		message += fmt.Sprintf("**Активные напоминания (%d):**\n\n", len(reminders))
		for i, r := range reminders {
			if i >= 5 { // Показываем только первые 5
				message += fmt.Sprintf("... и ещё %d\n", len(reminders)-5)
				break
			}
			dateTime := r.DateTime.Format("02.01.2006 15:04")
			message += fmt.Sprintf("• %s\n   📅 %s\n\n", r.Text, dateTime)
		}
	}

	keyboard := responder.NewKeyboardBuilder()
	row := keyboard.AddRow()
	row.AddCallback("➕ Создать напоминание", schemes.POSITIVE, "reminder:create")
	if len(reminders) > 0 {
		row2 := keyboard.AddRow()
		row2.AddCallback("📋 Все напоминания", schemes.POSITIVE, "reminder:list")
	}

	return responder.SendMarkdownWithKeyboard(ctx, req.Recipient(), message, keyboard)
}

func (h *ReminderHandler) HandleTextInput(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	if req.UserState == nil || req.UserState.UserRegistrationStep != "reminder_create" {
		return nil
	}

	text := strings.TrimSpace(req.Args)
	if text == "" {
		return responder.SendText(ctx, req.Recipient(), "❌ Текст не может быть пустым. Попробуй снова.")
	}

	if req.UserState.UserRegistrationData == nil {
		req.UserState.UserRegistrationData = make(map[string]string)
	}

	currentStep := req.UserState.UserRegistrationStep
	if currentStep == "reminder_create" {
		// Определяем, на каком подшаге мы находимся
		if req.UserState.UserRegistrationData["text"] == "" {
			// Шаг 1: Текст напоминания
			req.UserState.UserRegistrationData["text"] = text
			return h.showDateStep(ctx, req, responder)
		} else if req.UserState.UserRegistrationData["date"] == "" {
			// Шаг 2: Дата
			return h.handleDateInput(ctx, req, responder, text)
		} else if req.UserState.UserRegistrationData["time"] == "" {
			// Шаг 3: Время
			return h.handleTimeInput(ctx, req, responder, text)
		}
	}

	return nil
}

func (h *ReminderHandler) showDateStep(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := "✅ Текст напоминания сохранён.\n\n"
	message += "**Шаг 2 из 3: Выбери дату**\n\n"
	message += "Ты можешь выбрать быструю дату или ввести свою в формате ДД.ММ.ГГГГ:"

	now := time.Now()
	today := now.Format("02.01.2006")
	tomorrow := now.AddDate(0, 0, 1).Format("02.01.2006")
	weekLater := now.AddDate(0, 0, 7).Format("02.01.2006")

	keyboard := responder.NewKeyboardBuilder()
	row := keyboard.AddRow()
	row.AddCallback(fmt.Sprintf("📅 Сегодня (%s)", today), schemes.POSITIVE, fmt.Sprintf("reminder:date:%s", today))
	row2 := keyboard.AddRow()
	row2.AddCallback(fmt.Sprintf("📅 Завтра (%s)", tomorrow), schemes.POSITIVE, fmt.Sprintf("reminder:date:%s", tomorrow))
	row3 := keyboard.AddRow()
	row3.AddCallback(fmt.Sprintf("📅 Через неделю (%s)", weekLater), schemes.POSITIVE, fmt.Sprintf("reminder:date:%s", weekLater))
	row4 := keyboard.AddRow()
	row4.AddCallback("✏️ Ввести свою дату", schemes.POSITIVE, "reminder:date:custom")

	return responder.SendMarkdownWithKeyboard(ctx, req.Recipient(), message, keyboard)
}

func (h *ReminderHandler) handleDateInput(ctx context.Context, req *bot.Request, responder bot.Responder, dateStr string) error {
	// Парсим дату в формате ДД.ММ.ГГГГ
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		return responder.SendText(ctx, req.Recipient(), "❌ Неверный формат даты. Используй формат ДД.ММ.ГГГГ (например, 25.12.2024)")
	}

	// Проверяем, что дата не в прошлом
	now := time.Now()
	if date.Before(now.Truncate(24 * time.Hour)) {
		return responder.SendText(ctx, req.Recipient(), "❌ Нельзя создать напоминание на прошедшую дату. Выбери другую дату.")
	}

	req.UserState.UserRegistrationData["date"] = date.Format("02.01.2006")
	return h.showTimeStep(ctx, req, responder)
}

func (h *ReminderHandler) showTimeStep(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	message := "✅ Дата сохранена.\n\n"
	message += "**Шаг 3 из 3: Введи время**\n\n"
	message += "Введи время в формате ЧЧ:ММ (например, 14:30):"

	return responder.SendMarkdown(ctx, req.Recipient(), message)
}

func (h *ReminderHandler) handleTimeInput(ctx context.Context, req *bot.Request, responder bot.Responder, timeStr string) error {
	// Парсим время в формате ЧЧ:ММ
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 2 {
		return responder.SendText(ctx, req.Recipient(), "❌ Неверный формат времени. Используй формат ЧЧ:ММ (например, 14:30)")
	}

	hour, err := strconv.Atoi(timeParts[0])
	if err != nil || hour < 0 || hour > 23 {
		return responder.SendText(ctx, req.Recipient(), "❌ Неверный час. Используй значение от 0 до 23.")
	}

	minute, err := strconv.Atoi(timeParts[1])
	if err != nil || minute < 0 || minute > 59 {
		return responder.SendText(ctx, req.Recipient(), "❌ Неверная минута. Используй значение от 0 до 59.")
	}

	// Парсим дату
	dateStr := req.UserState.UserRegistrationData["date"]
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при обработке даты. Начни заново.")
	}

	// Создаем полную дату и время
	dateTime := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, time.Local)

	// Проверяем, что дата и время не в прошлом
	if dateTime.Before(time.Now()) {
		return responder.SendText(ctx, req.Recipient(), "❌ Нельзя создать напоминание на прошедшее время. Выбери другое время.")
	}

	// Создаем напоминание
	text := req.UserState.UserRegistrationData["text"]
	reminder, err := h.reminderService.CreateReminder(ctx, req.UserID(), text, dateTime)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", req.UserID()).Msg("failed to create reminder")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при создании напоминания")
	}

	// Очищаем состояние
	req.UserState.UserRegistrationStep = ""
	req.UserState.UserRegistrationData = nil

	message := fmt.Sprintf("✅ **Напоминание создано!**\n\n")
	message += fmt.Sprintf("**Текст:** %s\n", reminder.Text)
	message += fmt.Sprintf("**Дата и время:** %s\n\n", reminder.DateTime.Format("02.01.2006 15:04"))
	message += "Ты получишь напоминание в указанное время."

	return responder.SendMarkdown(ctx, req.Recipient(), message)
}

func (h *ReminderHandler) HandleCallback(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	payload := req.Args
	callbackID := ""

	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok {
			callbackID = cid
			responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
		}
	}

	userID := req.UserID()

	if payload == "reminder:create" {
		// Начинаем создание напоминания
		if req.UserState == nil {
			req.UserState = &state.UserState{}
		}
		req.UserState.UserRegistrationStep = "reminder_create"
		if req.UserState.UserRegistrationData == nil {
			req.UserState.UserRegistrationData = make(map[string]string)
		}

		message := "⏰ **Создание напоминания**\n\n"
		message += "**Шаг 1 из 3: Введи текст напоминания**\n\n"
		message += "Напиши, о чём тебе напомнить:"

		return responder.SendMarkdown(ctx, req.Recipient(), message)
	}

	if payload == "reminder:list" {
		// Показываем все напоминания
		reminders, err := h.reminderService.GetUserReminders(ctx, userID)
		if err != nil {
			h.logger.Error().Err(err).Str("user_id", userID).Msg("failed to get reminders")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении напоминаний")
		}

		if len(reminders) == 0 {
			return responder.SendMarkdown(ctx, req.Recipient(), "📋 У тебя пока нет напоминаний.")
		}

		message := fmt.Sprintf("📋 **Все напоминания (%d):**\n\n", len(reminders))
		for i, r := range reminders {
			if i >= 10 { // Ограничиваем до 10
				message += fmt.Sprintf("\n... и ещё %d", len(reminders)-10)
				break
			}
			status := "✅"
			if r.Status == "active" {
				status = "⏰"
			} else if r.Status == "cancelled" {
				status = "❌"
			}
			dateTime := r.DateTime.Format("02.01.2006 15:04")
			message += fmt.Sprintf("%s **%s**\n   📅 %s\n\n", status, r.Text, dateTime)
		}

		return responder.SendMarkdown(ctx, req.Recipient(), message)
	}

	if strings.HasPrefix(payload, "reminder:date:") {
		// Обработка выбора даты
		datePart := strings.TrimPrefix(payload, "reminder:date:")
		if datePart == "custom" {
			message := "✏️ **Введи дату**\n\n"
			message += "Введи дату в формате ДД.ММ.ГГГГ (например, 25.12.2024):"
			return responder.SendMarkdown(ctx, req.Recipient(), message)
		}

		// Парсим дату из callback
		date, err := time.Parse("02.01.2006", datePart)
		if err != nil {
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при обработке даты")
		}

		if req.UserState == nil {
			req.UserState = &state.UserState{}
		}
		if req.UserState.UserRegistrationData == nil {
			req.UserState.UserRegistrationData = make(map[string]string)
		}

		req.UserState.UserRegistrationStep = "reminder_create"
		req.UserState.UserRegistrationData["date"] = date.Format("02.01.2006")
		return h.showTimeStep(ctx, req, responder)
	}

	return nil
}

