package handlers

import (
	"context"
	"strings"

	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/ai"
	"first-max-bot/internal/services/moodle"
	"first-max-bot/internal/services/schedule"
	"first-max-bot/internal/services/user"
)

// AskHandler обрабатывает команду /ask для вопросов к AI
type AskHandler struct {
	aiService      ai.Service
	scheduleService schedule.Service
	moodleService  moodle.Service
	userService    user.Service
	logger         zerolog.Logger
}

func NewAskHandler(aiService ai.Service, scheduleService schedule.Service, moodleService moodle.Service, userService user.Service, logger zerolog.Logger) *AskHandler {
	return &AskHandler{
		aiService:       aiService,
		scheduleService: scheduleService,
		moodleService:   moodleService,
		userService:     userService,
		logger:          logger,
	}
}

func (h *AskHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Проверяем, что AI сервис инициализирован
	if h.aiService == nil {
		return responder.SendText(ctx, req.Recipient(), "❌ Сервис AI временно недоступен. Обратитесь к администратору.")
	}

	userID := req.UserID()
	question := strings.TrimSpace(req.Args)

	// Если вопрос пустой, просим задать вопрос
	if question == "" {
		message := "💬 **Задай вопрос**\n\n"
		message += "Я могу помочь тебе с вопросами о:\n"
		message += "• Расписании занятий\n"
		message += "• Курсах и обучении\n"
		message += "• Университетской жизни\n"
		message += "• И многом другом!\n\n"
		message += "Просто напиши свой вопрос после команды /ask."

		return responder.SendMarkdown(ctx, req.Recipient(), message)
	}

	// Получаем информацию о пользователе
	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil || u == nil {
		return responder.SendText(ctx, req.Recipient(), "❌ Пользователь не найден. Пожалуйста, зарегистрируйся через /register")
	}

	// Формируем контекстные данные
	contextData := ai.ContextData{
		UserInfo: ai.UserInfo{
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Age:       u.Age,
			Gender:    u.Gender,
			Email:     u.Email,
			Role:      h.getRoleLabel(u.Role),
		},
	}

	// Получаем расписание
	scheduleItems, err := h.scheduleService.GetSchedule(ctx, userID)
	if err == nil {
		for _, item := range scheduleItems {
			contextData.Schedule = append(contextData.Schedule, ai.ScheduleItem{
				Subject:  item.Discipline,
				Time:     item.Time.Format("15:04"),
				Date:     item.Time,
				Location: item.Location,
				Teacher:  item.Instructor,
			})
		}
	} else {
		h.logger.Warn().Err(err).Str("user_id", userID).Msg("failed to get schedule for AI context")
	}

	// Получаем курсы из Moodle (если есть токен)
	if u.MoodleToken != "" {
		siteInfo, err := h.moodleService.GetSiteInfo(ctx, u.MoodleToken)
		if err == nil {
			courses, err := h.moodleService.GetUserCourses(ctx, u.MoodleToken, siteInfo.UserID)
			if err == nil {
				for _, course := range courses {
					// Очищаем HTML из описания
					description := h.cleanHTML(course.Summary)
					if len(description) > 200 {
						description = description[:200] + "..."
					}

					contextData.Courses = append(contextData.Courses, ai.Course{
						Fullname:    course.Fullname,
						Description: description,
						StartDate:   course.StartDate,
						EndDate:     course.EndDate,
						Progress:    course.Progress,
						Completed:   course.Completed,
					})
				}
			} else {
				h.logger.Warn().Err(err).Str("user_id", userID).Msg("failed to get moodle courses for AI context")
			}
		} else {
			h.logger.Warn().Err(err).Str("user_id", userID).Msg("failed to get moodle site info for AI context")
		}
	}

	// Отправляем вопрос в AI
	response, err := h.aiService.AskQuestion(ctx, question, contextData)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Str("question", question).Msg("failed to get AI response")
		return responder.SendText(ctx, req.Recipient(), "❌ Извини, не удалось получить ответ. Попробуй позже.")
	}

	// Отправляем ответ пользователю
	return responder.SendMarkdown(ctx, req.Recipient(), response)
}

// getRoleLabel возвращает русское название роли
func (h *AskHandler) getRoleLabel(role user.Role) string {
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

// cleanHTML очищает HTML теги из текста
func (h *AskHandler) cleanHTML(html string) string {
	text := html
	// Убираем HTML теги (простая замена)
	text = strings.ReplaceAll(text, "<br />", "\n")
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "</h3>", "\n")
	text = strings.ReplaceAll(text, "</h5>", "\n")
	text = strings.ReplaceAll(text, "<h3>", "")
	text = strings.ReplaceAll(text, "<h5>", "")
	text = strings.ReplaceAll(text, "<strong>", "**")
	text = strings.ReplaceAll(text, "</strong>", "**")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	
	// Убираем все остальные HTML теги
	for strings.Contains(text, "<") && strings.Contains(text, ">") {
		start := strings.Index(text, "<")
		end := strings.Index(text[start:], ">")
		if end != -1 {
			text = text[:start] + text[start+end+1:]
		} else {
			break
		}
	}
	
	// Очищаем от лишних пробелов и переносов
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	
	return text
}

