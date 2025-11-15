package bot

import (
	"strings"

	"first-max-bot/internal/state"
)

type Router struct {
	handlers       map[string]Handler
	callbackRoutes map[string]Handler // маршруты для callback payloads
	fallback       Handler
}

func NewRouter() *Router {
	return &Router{
		handlers:       make(map[string]Handler),
		callbackRoutes: make(map[string]Handler),
	}
}

func (r *Router) Register(command string, handler Handler) {
	command = normalizeCommand(command)
	r.handlers[command] = handler
}

func (r *Router) SetFallback(handler Handler) {
	r.fallback = handler
}

func (r *Router) Resolve(text string) (Handler, string, string) {
	command, args := parseCommand(text)
	if h, ok := r.handlers[command]; ok {
		return h, command, args
	}
	return r.fallback, command, args
}

// ResolveByState разрешает handler на основе состояния пользователя
// Используется для обработки текстовых сообщений во время регистрации и ответов на обращения
func (r *Router) ResolveByState(text string, userState *state.UserState) (Handler, string, string) {
	if userState == nil {
		return r.Resolve(text)
	}

	// Если пользователь отвечает на обращение (проверяем ПЕРВЫМ, до регистрации)
	if userState.UserRegistrationStep == "ticket_reply" {
		if !strings.HasPrefix(strings.TrimSpace(text), "/") {
			if h, ok := r.handlers["/tickets"]; ok {
				return h, "", text // Передаем текст как args
			}
		}
	}

	// Если пользователь отвечает на ответ руководителя
	if userState.UserRegistrationStep == "ticket_user_reply" {
		if !strings.HasPrefix(strings.TrimSpace(text), "/") {
			if h, ok := r.handlers["/mytickets"]; ok {
				return h, "", text // Передаем текст как args
			}
		}
	}

	// Если администратор отвечает на заявление деканата
	if userState.UserRegistrationStep == "doc_response" {
		// Проверяем, что это не команда (не начинается с /)
		// Или это сообщение с файлом (даже если текст пустой)
		textTrimmed := strings.TrimSpace(text)
		if textTrimmed == "" || !strings.HasPrefix(textTrimmed, "/") {
			if h, ok := r.handlers["/documents"]; ok {
				return h, "", text // Передаем текст как args (может быть пустым для файла)
			}
		}
		// Если это команда, продолжаем обычную обработку ниже
	}

	// Если администратор отправляет новость
	if userState.UserRegistrationStep == "send_news" {
		// Если это не команда (не начинается с /), то это текстовый ввод для новости
		textTrimmed := strings.TrimSpace(text)
		if !strings.HasPrefix(textTrimmed, "/") {
			if h, ok := r.handlers["/send_news"]; ok {
				return h, "", text // Передаем текст как args
			}
		}
	}

	// Если пользователь вводит токен Moodle
	if userState.UserRegistrationStep == "moodle_token" {
		// Если это не команда (не начинается с /), то это текстовый ввод для токена
		textTrimmed := strings.TrimSpace(text)
		if !strings.HasPrefix(textTrimmed, "/") {
			if h, ok := r.handlers["/moodle"]; ok {
				return h, "", text // Передаем текст как args
			}
		}
	}

	// Если пользователь создает напоминание
	if userState.UserRegistrationStep == "reminder_create" {
		// Если это не команда (не начинается с /), то это текстовый ввод для напоминания
		textTrimmed := strings.TrimSpace(text)
		if !strings.HasPrefix(textTrimmed, "/") {
			if h, ok := r.handlers["/reminder"]; ok {
				return h, "", text // Передаем текст как args
			}
		}
	}

	// Если пользователь в процессе регистрации
	if userState.UserRegistrationStep != "" && userState.UserRegistrationStep != "completed" && userState.UserRegistrationStep != "ticket_reply" && userState.UserRegistrationStep != "ticket_user_reply" && userState.UserRegistrationStep != "doc_response" && userState.UserRegistrationStep != "send_news" && userState.UserRegistrationStep != "moodle_token" && userState.UserRegistrationStep != "reminder_create" {
		// Если это не команда (не начинается с /), то это текстовый ввод для регистрации
		if !strings.HasPrefix(strings.TrimSpace(text), "/") {
			if h, ok := r.handlers["/register"]; ok {
				return h, "", text // Передаем текст как args
			}
		}
	}
	
	// Обычная логика разрешения
	return r.Resolve(text)
}

func (r *Router) RegisterCallback(payload string, handler Handler) {
	r.callbackRoutes[payload] = handler
}

func (r *Router) ResolveCallback(payload string, userState *state.UserState) Handler {
	// Сначала проверяем точное совпадение payload
	if h, ok := r.callbackRoutes[payload]; ok {
		return h
	}

	// Проверяем общие handlers по префиксам
	// Маппинг префикса на wildcard pattern
	prefixToWildcard := map[string]string{
		"user_reg:":   "user_reg:*",
		"ticket:":     "ticket:*",
		"myticket:":   "myticket:*",
		"doc:":        "doc:*",
		"doc_admin:":  "doc_admin:*",
		"book:":       "book:*",
		"lib_manage:": "lib_manage:*",
		"moodle:":     "moodle:*",
		"reminder:":   "reminder:*",
	}

	for prefix, wildcard := range prefixToWildcard {
		if strings.HasPrefix(payload, prefix) {
			if h, ok := r.callbackRoutes[wildcard]; ok {
				return h
			}
		}
	}

	return nil
}

func parseCommand(text string) (string, string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ""
	}

	fields := strings.Fields(text)
	command := normalizeCommand(fields[0])
	args := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	return command, args
}

func normalizeCommand(command string) string {
	command = strings.TrimSpace(strings.ToLower(command))
	if command == "" {
		return command
	}
	if command[0] == '/' {
		return command
	}
	return "/" + command
}
