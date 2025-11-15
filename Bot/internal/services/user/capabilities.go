package user

// Capability представляет возможность/функцию, доступную пользователю
type Capability string

const (
	// Общие возможности
	CapabilitySchedule  Capability = "schedule"   // Расписание
	CapabilityContact   Capability = "contact"    // Обращение в поддержку
	CapabilityMyTickets Capability = "my_tickets" // Мои обращения
	CapabilityHelp      Capability = "help"       // Справка
	CapabilityReminder  Capability = "reminder"   // Напоминания
	CapabilityAsk       Capability = "ask"        // Вопросы к AI

	// Возможности для абитуриентов
	CapabilityAdmissionInfo Capability = "admission_info" // Информация о поступлении
	CapabilityPrograms      Capability = "programs"       // Программы обучения
	CapabilityOpenDay       Capability = "open_day"       // День открытых дверей

	// Возможности для студентов
	CapabilityStudentSchedule Capability = "student_schedule" // Расписание занятий
	CapabilityDeanery         Capability = "deanery"          // Деканат (справки, заявления)
	CapabilityLibrary         Capability = "library"          // Библиотека
	CapabilityDormitory       Capability = "dormitory"        // Общежитие
	CapabilityCareer          Capability = "career"           // Карьера
	CapabilityProjects        Capability = "projects"         // Проектная деятельность
	CapabilityEvents          Capability = "events"           // Внеучебная деятельность
	CapabilityMoodle          Capability = "moodle"           // Moodle интеграция

	// Возможности для сотрудников
	CapabilityBusinessTrip  Capability = "business_trip"  // Командировки
	CapabilityVacation      Capability = "vacation"       // Отпуска
	CapabilityOffice        Capability = "office"         // Офис (справки, пропуски)
	CapabilityLibraryManage Capability = "library_manage" // Управление библиотекой

	// Возможности для руководителей
	CapabilityDashboard Capability = "dashboard" // Дашборд
	CapabilityAnalytics Capability = "analytics" // Аналитика
	CapabilityNews      Capability = "news"      // Новости
	CapabilitySendNews  Capability = "send_news" // Отправка новостей
	CapabilityTickets   Capability = "tickets"   // Управление обращениями
	CapabilityDocuments Capability = "documents" // Заявления деканата
)

// RoleCapabilities определяет возможности для каждой роли
var RoleCapabilities = map[Role][]Capability{
	RoleApplicant: {
		CapabilityHelp,
		CapabilityAdmissionInfo,
		CapabilityPrograms,
		CapabilityOpenDay,
		CapabilityReminder,
		CapabilityAsk,
	},
	RoleStudent: {
		CapabilityHelp,
		CapabilitySchedule,
		CapabilityStudentSchedule,
		CapabilityDeanery,
		CapabilityLibrary,
		CapabilityDormitory,
		CapabilityCareer,
		CapabilityProjects,
		CapabilityEvents,
		CapabilityMoodle,
		CapabilityContact,
		CapabilityMyTickets,
		CapabilityReminder,
		CapabilityAsk,
	},
	RoleEmployee: {
		CapabilityHelp,
		CapabilitySchedule,
		CapabilityBusinessTrip,
		CapabilityVacation,
		CapabilityOffice,
		CapabilityContact,
		CapabilityLibraryManage,
		CapabilityReminder,
		CapabilityAsk,
	},
	RoleManager: {
		CapabilityHelp,
		CapabilitySchedule,
		CapabilityDashboard,
		CapabilityAnalytics,
		CapabilityNews,
		CapabilitySendNews,
		CapabilityTickets,
		CapabilityContact,
		CapabilityDocuments,
		CapabilityReminder,
		CapabilityAsk,
	},
}

// GetCapabilities возвращает список возможностей для роли
func GetCapabilities(role Role) []Capability {
	if caps, ok := RoleCapabilities[role]; ok {
		return caps
	}
	// По умолчанию возвращаем базовые возможности
	return []Capability{CapabilityHelp}
}

// HasCapability проверяет, есть ли у роли определенная возможность
func HasCapability(role Role, capability Capability) bool {
	caps := GetCapabilities(role)
	for _, cap := range caps {
		if cap == capability {
			return true
		}
	}
	return false
}

// CommandInfo содержит информацию о команде
type CommandInfo struct {
	Command     string
	Description string
	Capability  Capability
}

// GetCommandsForRole возвращает список команд для роли
func GetCommandsForRole(role Role) []CommandInfo {
	caps := GetCapabilities(role)
	commands := make([]CommandInfo, 0, len(caps))

	for _, cap := range caps {
		cmd := getCommandForCapability(cap)
		if cmd.Command != "" {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// getCommandForCapability возвращает информацию о команде для возможности
func getCommandForCapability(cap Capability) CommandInfo {
	switch cap {
	case CapabilityHelp:
		return CommandInfo{Command: "/help", Description: "Справка по командам", Capability: cap}
	case CapabilitySchedule:
		return CommandInfo{Command: "/schedule", Description: "Расписание", Capability: cap}
	case CapabilityContact:
		return CommandInfo{Command: "/contact", Description: "Обращение в поддержку", Capability: cap}
	case CapabilityMyTickets:
		return CommandInfo{Command: "/mytickets", Description: "Мои обращения", Capability: cap}
	case CapabilityAdmissionInfo:
		return CommandInfo{Command: "/admission", Description: "Информация о поступлении", Capability: cap}
	case CapabilityPrograms:
		return CommandInfo{Command: "/programs", Description: "Программы обучения", Capability: cap}
	case CapabilityOpenDay:
		return CommandInfo{Command: "/openday", Description: "День открытых дверей", Capability: cap}
	case CapabilityStudentSchedule:
		return CommandInfo{Command: "/myschedule", Description: "Моё расписание", Capability: cap}
	case CapabilityDeanery:
		return CommandInfo{Command: "/deanery", Description: "Деканат", Capability: cap}
	case CapabilityLibrary:
		return CommandInfo{Command: "/library", Description: "Библиотека", Capability: cap}
	case CapabilityDormitory:
		return CommandInfo{Command: "/dormitory", Description: "Общежитие", Capability: cap}
	case CapabilityMoodle:
		return CommandInfo{Command: "/moodle", Description: "Moodle", Capability: cap}
	case CapabilityOffice:
		return CommandInfo{Command: "/office", Description: "Офис", Capability: cap}
	case CapabilityLibraryManage:
		return CommandInfo{Command: "/library_manage", Description: "Управление библиотекой", Capability: cap}
	case CapabilityDashboard:
		return CommandInfo{Command: "/dashboard", Description: "Дашборд", Capability: cap}
	case CapabilityAnalytics:
		return CommandInfo{Command: "/analytics", Description: "Аналитика", Capability: cap}
	case CapabilityNews:
		return CommandInfo{Command: "/news", Description: "Новости", Capability: cap}
	case CapabilitySendNews:
		return CommandInfo{Command: "/send_news", Description: "Отправить новость", Capability: cap}
	case CapabilityTickets:
		return CommandInfo{Command: "/tickets", Description: "Управление обращениями", Capability: cap}
	case CapabilityDocuments:
		return CommandInfo{Command: "/documents", Description: "Заявления деканата", Capability: cap}
	case CapabilityReminder:
		return CommandInfo{Command: "/reminder", Description: "Напоминания", Capability: cap}
	case CapabilityAsk:
		return CommandInfo{Command: "/ask", Description: "Задать вопрос", Capability: cap}
	default:
		return CommandInfo{}
	}
}
