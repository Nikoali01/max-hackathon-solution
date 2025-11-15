package handlers

import (
	"context"
	"fmt"
	"strings"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/user"
)

type MenuHandler struct {
	userService user.Service
}

func NewMenuHandler(userService user.Service) *MenuHandler {
	return &MenuHandler{
		userService: userService,
	}
}

func (h *MenuHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userID := req.UserID()
	if userID == "" {
		return responder.SendText(ctx, req.Recipient(), "Не удалось определить пользователя")
	}

	// Получаем пользователя
	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении данных пользователя")
	}

	if u == nil {
		return responder.SendText(ctx, req.Recipient(), "❌ Ты не зарегистрирован. Используй /register для регистрации.")
	}

	// Получаем команды для роли
	commands := user.GetCommandsForRole(u.Role)
	if len(commands) == 0 {
		return responder.SendText(ctx, req.Recipient(), "❌ Нет доступных команд для твоей роли.")
	}

	// Формируем меню
	var builder strings.Builder
	builder.WriteString("📋 Доступные команды:\n\n")
	
	// Группируем команды по категориям
	roleLabel := h.getRoleLabel(u.Role)
	builder.WriteString(fmt.Sprintf("Роль: %s\n\n", roleLabel))

	// Общие команды (для всех ролей)
	generalCaps := map[user.Capability]bool{
		user.CapabilityHelp:    true,
		user.CapabilitySchedule: true,
		user.CapabilityContact: true,
	}
	
	generalCommands := []user.CommandInfo{}
	roleSpecificCommands := []user.CommandInfo{}
	
	for _, cmd := range commands {
		if generalCaps[cmd.Capability] {
			generalCommands = append(generalCommands, cmd)
		} else {
			roleSpecificCommands = append(roleSpecificCommands, cmd)
		}
	}

	// Общие команды
	if len(generalCommands) > 0 {
		builder.WriteString("🔹 Общее:\n")
		for _, cmd := range generalCommands {
			builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
		}
		builder.WriteString("\n")
	}

	// Команды в зависимости от роли
	switch u.Role {
	case user.RoleApplicant:
		if len(roleSpecificCommands) > 0 {
			builder.WriteString("🔹 Для абитуриентов:\n")
			for _, cmd := range roleSpecificCommands {
				builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
			}
		}
	case user.RoleStudent:
		if len(roleSpecificCommands) > 0 {
			builder.WriteString("🔹 Для студентов:\n")
			// Группируем по категориям
			studyCommands := []user.CommandInfo{}
			serviceCommands := []user.CommandInfo{}
			activityCommands := []user.CommandInfo{}
			
			for _, cmd := range roleSpecificCommands {
				switch cmd.Capability {
				case user.CapabilityStudentSchedule, user.CapabilityDeanery, user.CapabilityLibrary, user.CapabilityDormitory:
					studyCommands = append(studyCommands, cmd)
				case user.CapabilityMyTickets:
					serviceCommands = append(serviceCommands, cmd)
				case user.CapabilityCareer, user.CapabilityProjects, user.CapabilityEvents:
					activityCommands = append(activityCommands, cmd)
				default:
					studyCommands = append(studyCommands, cmd)
				}
			}
			
			if len(studyCommands) > 0 {
				for _, cmd := range studyCommands {
					builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
				}
			}
			if len(serviceCommands) > 0 {
				for _, cmd := range serviceCommands {
					builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
				}
			}
			if len(activityCommands) > 0 {
				for _, cmd := range activityCommands {
					builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
				}
			}
		}
	case user.RoleEmployee:
		if len(roleSpecificCommands) > 0 {
			builder.WriteString("🔹 Для сотрудников:\n")
			// Группируем по категориям
			workCommands := []user.CommandInfo{}
			manageCommands := []user.CommandInfo{}
			serviceCommands := []user.CommandInfo{}
			
			for _, cmd := range roleSpecificCommands {
				switch cmd.Capability {
				case user.CapabilityBusinessTrip, user.CapabilityVacation, user.CapabilityOffice:
					workCommands = append(workCommands, cmd)
				case user.CapabilityLibraryManage:
					manageCommands = append(manageCommands, cmd)
				default:
					serviceCommands = append(serviceCommands, cmd)
				}
			}
			
			if len(workCommands) > 0 {
				for _, cmd := range workCommands {
					builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
				}
			}
			if len(manageCommands) > 0 {
				for _, cmd := range manageCommands {
					builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
				}
			}
			if len(serviceCommands) > 0 {
				for _, cmd := range serviceCommands {
					builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
				}
			}
		}
	case user.RoleManager:
		if len(roleSpecificCommands) > 0 {
			builder.WriteString("🔹 Для руководителей:\n")
			// Группируем по категориям
			analyticsCommands := []user.CommandInfo{}
			newsCommands := []user.CommandInfo{}
			manageCommands := []user.CommandInfo{}
			
			for _, cmd := range roleSpecificCommands {
				switch cmd.Capability {
				case user.CapabilityDashboard, user.CapabilityAnalytics:
					analyticsCommands = append(analyticsCommands, cmd)
				case user.CapabilityNews, user.CapabilitySendNews:
					newsCommands = append(newsCommands, cmd)
				case user.CapabilityTickets, user.CapabilityDocuments:
					manageCommands = append(manageCommands, cmd)
				default:
					manageCommands = append(manageCommands, cmd)
				}
			}
			
			if len(analyticsCommands) > 0 {
				for _, cmd := range analyticsCommands {
					builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
				}
			}
			if len(newsCommands) > 0 {
				for _, cmd := range newsCommands {
					builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
				}
			}
			if len(manageCommands) > 0 {
				for _, cmd := range manageCommands {
					builder.WriteString(fmt.Sprintf("  %s — %s\n", cmd.Command, cmd.Description))
				}
			}
		}
	}

	builder.WriteString("\n💡 Используй команды для взаимодействия с ботом.")

	return responder.SendText(ctx, req.Recipient(), builder.String())
}

func (h *MenuHandler) getRoleLabel(role user.Role) string {
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

