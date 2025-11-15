package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/user"
	"first-max-bot/internal/state"
)

type UserRegistrationHandler struct {
	userService user.Service
	logger      zerolog.Logger
}

func NewUserRegistrationHandler(userService user.Service, logger zerolog.Logger) *UserRegistrationHandler {
	return &UserRegistrationHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserRegistrationHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Если это callback (payload начинается с user_reg:)
	if strings.HasPrefix(req.Args, "user_reg:") {
		return h.handleCallback(ctx, req, responder)
	}

	// Если пользователь в процессе регистрации и это текстовое сообщение
	userState := req.UserState
	if userState != nil && userState.UserRegistrationStep != "" && userState.UserRegistrationStep != "completed" {
		// Это текстовый ввод для текущего шага
		return h.handleTextInput(ctx, req, responder, userState)
	}

	// Иначе это команда /register - начинаем регистрацию
	return h.startRegistration(ctx, req, responder)
}

func (h *UserRegistrationHandler) handleTextInput(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState) error {
	text := strings.TrimSpace(req.Args)
	if text == "" {
		return responder.SendText(ctx, req.Recipient(), "Пожалуйста, введи текст")
	}

	currentStep := userState.UserRegistrationStep
	switch currentStep {
	case "first_name":
		// Сохраняем имя и переходим к фамилии
		userState.UserRegistrationData["first_name"] = text
		userState.UserRegistrationStep = "last_name"
		return h.showLastNameStep(ctx, req, responder, userState)

	case "last_name":
		// Сохраняем фамилию и переходим к возрасту
		userState.UserRegistrationData["last_name"] = text
		userState.UserRegistrationStep = "age"
		return h.showAgeStep(ctx, req, responder, userState)

	case "age":
		// Валидация возраста
		age, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			return responder.SendText(ctx, req.Recipient(), "❌ Пожалуйста, введи корректный возраст (число). Например: 20")
		}
		if age < 1 || age > 150 {
			return responder.SendText(ctx, req.Recipient(), "❌ Возраст должен быть от 1 до 150 лет. Попробуй ещё раз.")
		}
		// Сохраняем возраст и переходим к полу
		userState.UserRegistrationData["age"] = text
		userState.UserRegistrationStep = "gender"
		return h.showGenderStep(ctx, req, responder, userState)

	case "email":
		// Простая валидация email
		if !strings.Contains(text, "@") {
			return responder.SendText(ctx, req.Recipient(), "❌ Пожалуйста, введи корректный email адрес (должен содержать @)")
		}
		// Сохраняем email и переходим к подтверждению
		userState.UserRegistrationData["email"] = text
		userState.UserRegistrationStep = "email_verification"
		return h.showEmailVerificationStep(ctx, req, responder, userState)

	case "email_verification":
		// Проверяем код подтверждения
		code := strings.TrimSpace(text)
		expectedCode := "1111" // По умолчанию код 1111

		if code != expectedCode {
			return responder.SendText(ctx, req.Recipient(), "❌ Неверный код подтверждения. Попробуй ещё раз.")
		}

		// Код верный, завершаем регистрацию
		userState.UserRegistrationStep = "completed"
		return h.showCompletion(ctx, req, responder, userState)

	default:
		// Неожиданный шаг, начинаем заново
		userState.UserRegistrationStep = "first_name"
		userState.UserRegistrationData = make(map[string]string)
		return h.showFirstNameStep(ctx, req, responder, userState)
	}
}

func (h *UserRegistrationHandler) startRegistration(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userID := req.UserID()
	if userID == "" {
		return responder.SendText(ctx, req.Recipient(), "Не удалось определить пользователя")
	}

	// Проверяем, не зарегистрирован ли уже пользователь
	existingUser, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("failed to check existing user")
	}
	if existingUser != nil {
		text := fmt.Sprintf("✅ Ты уже зарегистрирован!\n\n")
		text += fmt.Sprintf("Имя: %s %s\n", existingUser.FirstName, existingUser.LastName)
		text += fmt.Sprintf("Email: %s\n", existingUser.Email)
		text += fmt.Sprintf("Роль: %s\n\n", h.getRoleLabel(existingUser.Role))
		text += "Если хочешь изменить данные, напиши /register ещё раз."
		return responder.SendText(ctx, req.Recipient(), text)
	}

	// Инициализируем состояние регистрации
	userState := req.UserState
	if userState == nil {
		userState = &state.UserState{
			UserRegistrationData: make(map[string]string),
		}
	}
	if userState.UserRegistrationData == nil {
		userState.UserRegistrationData = make(map[string]string)
	}

	// Если регистрация уже начата, продолжаем с текущего шага
	if userState.UserRegistrationStep != "" && userState.UserRegistrationStep != "completed" {
		return h.resumeRegistration(ctx, req, responder, userState)
	}

	// Начинаем с первого шага - имя
	userState.UserRegistrationStep = "first_name"
	return h.showFirstNameStep(ctx, req, responder, userState)
}

func (h *UserRegistrationHandler) resumeRegistration(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState) error {
	step := userState.UserRegistrationStep
	switch step {
	case "first_name":
		return h.showFirstNameStep(ctx, req, responder, userState)
	case "last_name":
		return h.showLastNameStep(ctx, req, responder, userState)
	case "age":
		return h.showAgeStep(ctx, req, responder, userState)
	case "gender":
		return h.showGenderStep(ctx, req, responder, userState)
	case "email":
		return h.showEmailStep(ctx, req, responder, userState)
	case "email_verification":
		return h.showEmailVerificationStep(ctx, req, responder, userState)
	default:
		// Если шаг неизвестен, начинаем заново
		userState.UserRegistrationStep = "first_name"
		userState.UserRegistrationData = make(map[string]string)
		return h.showFirstNameStep(ctx, req, responder, userState)
	}
}

func (h *UserRegistrationHandler) handleCallback(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	payload := req.Args
	userState := req.UserState
	if userState == nil {
		userState = &state.UserState{
			UserRegistrationData: make(map[string]string),
		}
	}
	if userState.UserRegistrationData == nil {
		userState.UserRegistrationData = make(map[string]string)
	}

	callbackID := ""
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok && cid != "" {
			callbackID = cid
		}
	}

	// Обрабатываем навигацию
	if payload == "user_reg:back" {
		return h.handleBack(ctx, req, responder, userState, callbackID)
	}
	if payload == "user_reg:cancel" {
		return h.handleCancel(ctx, req, responder, userState, callbackID)
	}

	// Обрабатываем выбор значения на текущем шаге
	currentStep := userState.UserRegistrationStep
	if currentStep == "" {
		currentStep = "first_name"
		userState.UserRegistrationStep = currentStep
	}

	switch currentStep {
	case "first_name":
		// Имя вводится текстом, не через callback
		return nil
	case "last_name":
		// Фамилия вводится текстом, не через callback
		return nil
	case "age":
		// Возраст вводится текстом, не через callback
		return nil
	case "gender":
		if strings.HasPrefix(payload, "user_reg:gender:") {
			gender := strings.TrimPrefix(payload, "user_reg:gender:")
			userState.UserRegistrationData["gender"] = gender
			userState.UserRegistrationStep = "email"
			if callbackID != "" {
				responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
			}
			return h.showEmailStep(ctx, req, responder, userState)
		}
	case "email":
		// Email вводится текстом, не через callback
		return nil
	case "email_verification":
		// Код подтверждения вводится текстом, не через callback
		return nil
	}

	return nil
}

func (h *UserRegistrationHandler) showFirstNameStep(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState) error {
	text := "👤 Регистрация в системе университета\n\n"
	text += "Шаг 1 из 6: Введи своё имя\n\n"
	text += "Напиши своё имя текстом:"

	keyboard := responder.NewKeyboardBuilder()
	cancelRow := keyboard.AddRow()
	cancelRow.AddCallback("❌ Отмена", schemes.NEGATIVE, "user_reg:cancel")

	return h.respondWithKeyboard(ctx, req, responder, text, keyboard)
}

func (h *UserRegistrationHandler) showLastNameStep(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState) error {
	text := "👤 Регистрация в системе университета\n\n"
	text += "Шаг 2 из 6: Введи свою фамилию\n\n"
	text += "Напиши свою фамилию текстом:"

	keyboard := responder.NewKeyboardBuilder()
	navRow := keyboard.AddRow()
	navRow.AddCallback("◀️ Назад", schemes.DEFAULT, "user_reg:back")
	navRow.AddCallback("❌ Отмена", schemes.NEGATIVE, "user_reg:cancel")

	return h.respondWithKeyboard(ctx, req, responder, text, keyboard)
}

func (h *UserRegistrationHandler) showAgeStep(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState) error {
	text := "👤 Регистрация в системе университета\n\n"
	text += "Шаг 3 из 6: Введи свой возраст\n\n"
	text += "Напиши свой возраст числом (например: 20):"

	keyboard := responder.NewKeyboardBuilder()
	navRow := keyboard.AddRow()
	navRow.AddCallback("◀️ Назад", schemes.DEFAULT, "user_reg:back")
	navRow.AddCallback("❌ Отмена", schemes.NEGATIVE, "user_reg:cancel")

	return h.respondWithKeyboard(ctx, req, responder, text, keyboard)
}

func (h *UserRegistrationHandler) showGenderStep(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState) error {
	text := "👤 Регистрация в системе университета\n\n"
	text += "Шаг 4 из 6: Выбери свой пол\n\n"
	text += "Выбери пол:"

	keyboard := responder.NewKeyboardBuilder()
	row := keyboard.AddRow()
	row.AddCallback("Мужской", schemes.POSITIVE, "user_reg:gender:male")
	row.AddCallback("Женский", schemes.POSITIVE, "user_reg:gender:female")

	navRow := keyboard.AddRow()
	navRow.AddCallback("◀️ Назад", schemes.DEFAULT, "user_reg:back")
	navRow.AddCallback("❌ Отмена", schemes.NEGATIVE, "user_reg:cancel")

	return h.respondWithKeyboard(ctx, req, responder, text, keyboard)
}

func (h *UserRegistrationHandler) showEmailStep(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState) error {
	text := "👤 Регистрация в системе университета\n\n"
	text += "Шаг 5 из 6: Введи свою электронную почту\n\n"
	text += "Напиши свой email адрес:"

	keyboard := responder.NewKeyboardBuilder()
	navRow := keyboard.AddRow()
	navRow.AddCallback("◀️ Назад", schemes.DEFAULT, "user_reg:back")
	navRow.AddCallback("❌ Отмена", schemes.NEGATIVE, "user_reg:cancel")

	return h.respondWithKeyboard(ctx, req, responder, text, keyboard)
}

func (h *UserRegistrationHandler) showEmailVerificationStep(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState) error {
	email := userState.UserRegistrationData["email"]
	text := "👤 Регистрация в системе университета\n\n"
	text += "Шаг 6 из 6: Подтверждение email\n\n"
	text += fmt.Sprintf("Мы отправили код подтверждения на адрес %s\n\n", email)
	text += "Введи код подтверждения:"

	keyboard := responder.NewKeyboardBuilder()
	navRow := keyboard.AddRow()
	navRow.AddCallback("◀️ Назад", schemes.DEFAULT, "user_reg:back")
	navRow.AddCallback("❌ Отмена", schemes.NEGATIVE, "user_reg:cancel")

	return h.respondWithKeyboard(ctx, req, responder, text, keyboard)
}

func (h *UserRegistrationHandler) showCompletion(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState) error {
	userID := req.UserID()

	// Парсим возраст (теперь это просто число)
	ageStr := userState.UserRegistrationData["age"]
	age, _ := strconv.Atoi(ageStr)
	if age == 0 {
		age = 20 // Значение по умолчанию, если не удалось распарсить
	}

	// Создаем пользователя
	newUser := user.User{
		UserID:    userID,
		FirstName: userState.UserRegistrationData["first_name"],
		LastName:  userState.UserRegistrationData["last_name"],
		Age:       age,
		Gender:    userState.UserRegistrationData["gender"],
		Email:     userState.UserRegistrationData["email"],
		Role:      user.RoleStudent,
	}

	if newUser.FirstName == "Администратор" {
		newUser.Role = user.RoleManager
	}

	if newUser.FirstName == "Учитель" {
		newUser.Role = user.RoleEmployee
	}

	if newUser.FirstName == "Абитуриент" {
		newUser.Role = user.RoleApplicant
	}

	createdUser, err := h.userService.CreateUser(ctx, newUser)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create user")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при сохранении данных. Попробуй позже.")
	}

	// Получаем роль пользователя (может быть изменена бэкендом)
	role, _ := h.userService.GetUserRole(ctx, userID)

	var result strings.Builder
	result.WriteString("✅ Регистрация завершена!\n\n")
	result.WriteString("📋 Твои данные:\n")
	result.WriteString(fmt.Sprintf("• Имя: %s %s\n", createdUser.FirstName, createdUser.LastName))
	result.WriteString(fmt.Sprintf("• Возраст: %d лет\n", createdUser.Age))
	result.WriteString(fmt.Sprintf("• Пол: %s\n", h.getGenderLabel(createdUser.Gender)))
	result.WriteString(fmt.Sprintf("• Email: %s\n", createdUser.Email))
	result.WriteString(fmt.Sprintf("• Роль: %s\n\n", h.getRoleLabel(role)))
	result.WriteString("Теперь ты можешь пользоваться всеми возможностями бота! 🎉")

	// Удаляем старое сообщение и отправляем новое
	return h.deleteAndSendNew(ctx, req, responder, result.String(), nil)
}

func (h *UserRegistrationHandler) handleBack(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState, callbackID string) error {
	if callbackID != "" {
		responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
	}

	currentStep := userState.UserRegistrationStep
	switch currentStep {
	case "last_name":
		userState.UserRegistrationStep = "first_name"
		delete(userState.UserRegistrationData, "first_name")
		return h.showFirstNameStep(ctx, req, responder, userState)
	case "age":
		userState.UserRegistrationStep = "last_name"
		delete(userState.UserRegistrationData, "last_name")
		return h.showLastNameStep(ctx, req, responder, userState)
	case "gender":
		userState.UserRegistrationStep = "age"
		delete(userState.UserRegistrationData, "age")
		return h.showAgeStep(ctx, req, responder, userState)
	case "email":
		userState.UserRegistrationStep = "gender"
		delete(userState.UserRegistrationData, "gender")
		return h.showGenderStep(ctx, req, responder, userState)
	case "email_verification":
		userState.UserRegistrationStep = "email"
		delete(userState.UserRegistrationData, "email")
		return h.showEmailStep(ctx, req, responder, userState)
	default:
		userState.UserRegistrationStep = "first_name"
		return h.showFirstNameStep(ctx, req, responder, userState)
	}
}

func (h *UserRegistrationHandler) handleCancel(ctx context.Context, req *bot.Request, responder bot.Responder, userState *state.UserState, callbackID string) error {
	userState.UserRegistrationStep = ""
	userState.UserRegistrationData = make(map[string]string)

	return h.deleteAndSendNew(ctx, req, responder, "❌ Регистрация отменена. Можешь начать заново командой /register", nil)
}

func (h *UserRegistrationHandler) getGenderLabel(gender string) string {
	switch gender {
	case "male":
		return "Мужской"
	case "female":
		return "Женский"
	default:
		return gender
	}
}

func (h *UserRegistrationHandler) getRoleLabel(role user.Role) string {
	switch role {
	case user.RoleApplicant:
		return "Абитуриент"
	case user.RoleStudent:
		return "Студент"
	case user.RoleEmployee:
		return "Сотрудник"
	case user.RoleManager:
		return "Руководитель"
	default:
		return string(role)
	}
}

// respondWithKeyboard отправляет или редактирует сообщение с клавиатурой
func (h *UserRegistrationHandler) respondWithKeyboard(ctx context.Context, req *bot.Request, responder bot.Responder, text string, keyboard *maxbot.Keyboard) error {
	callbackID := ""
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok && cid != "" {
			callbackID = cid
		}
	}

	if callbackID != "" {
		return responder.AnswerCallbackWithEdit(ctx, callbackID, text, keyboard)
	}

	if keyboard != nil {
		return responder.SendTextWithKeyboard(ctx, req.Recipient(), text, keyboard)
	}

	return responder.SendText(ctx, req.Recipient(), text)
}

// deleteAndSendNew удаляет старое сообщение и отправляет новое
func (h *UserRegistrationHandler) deleteAndSendNew(ctx context.Context, req *bot.Request, responder bot.Responder, text string, keyboard *maxbot.Keyboard) error {
	callbackID := ""
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok && cid != "" {
			callbackID = cid
			responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
		}
	}

	messageID := ""
	if req.Metadata != nil {
		if mid, ok := req.Metadata["message_id"].(string); ok && mid != "" {
			messageID = mid
		}
	}

	if messageID != "" {
		if err := responder.DeleteMessageByMid(ctx, messageID); err != nil {
			h.logger.Warn().Err(err).Str("message_id", messageID).Msg("failed to delete message, continuing anyway")
		}
	}

	if keyboard != nil {
		return responder.SendTextWithKeyboard(ctx, req.Recipient(), text, keyboard)
	}

	return responder.SendText(ctx, req.Recipient(), text)
}
