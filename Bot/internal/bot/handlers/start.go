package handlers

import (
	"context"
	"fmt"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/user"
)

type StartHandler struct {
	userService user.Service
}

func NewStartHandler(userService user.Service, logger interface{}) *StartHandler {
	return &StartHandler{
		userService: userService,
	}
}

func (h *StartHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userID := req.UserID()
	
	// Проверяем, зарегистрирован ли пользователь
	existingUser, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		// В случае ошибки просто показываем приветствие
		return h.showWelcome(ctx, responder, req, nil)
	}
	
	if existingUser == nil {
		// Пользователь не зарегистрирован
		message := `👋 Привет! Я MAX Helper — твой ассистент для поступления и учебы.

Для начала работы нужно пройти регистрацию. Это займет всего пару минут!

Нажми /register чтобы начать регистрацию.`
		return responder.SendText(ctx, req.Recipient(), message)
	}
	
	// Пользователь зарегистрирован - показываем приветствие с его данными
	return h.showWelcome(ctx, responder, req, existingUser)
}

func (h *StartHandler) showWelcome(ctx context.Context, responder bot.Responder, req *bot.Request, u *user.User) error {
	var message string
	
	if u != nil {
		roleLabel := h.getRoleLabel(u.Role)
		message = fmt.Sprintf(`👋 Привет, %s %s!

Твоя роль: %s

Вот чем я могу помочь:`, u.FirstName, u.LastName, roleLabel)
		
		// Показываем команды в зависимости от роли
		commands := user.GetCommandsForRole(u.Role)
		for _, cmd := range commands {
			message += fmt.Sprintf("\n• %s — %s", cmd.Command, cmd.Description)
		}
	} else {
		message = `👋 Привет! Я MAX Helper — твой ассистент для поступления и учебы.

Для начала работы нужно пройти регистрацию. Это займет всего пару минут!

Нажми /register чтобы начать регистрацию.`
	}
	
	message += "\n\nНапиши команду или используй /menu для просмотра всех доступных команд."
	
	return responder.SendText(ctx, req.Recipient(), message)
}

func (h *StartHandler) getRoleLabel(role user.Role) string {
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
