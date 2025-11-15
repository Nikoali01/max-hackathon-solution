package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	botpkg "first-max-bot/internal/bot"
	"first-max-bot/internal/bot/handlers"
	"first-max-bot/internal/config"
	"first-max-bot/internal/services/ai"
	"first-max-bot/internal/services/businesstrip"
	"first-max-bot/internal/services/deanery"
	"first-max-bot/internal/services/library"
	"first-max-bot/internal/services/moodle"
	"first-max-bot/internal/services/news"
	"first-max-bot/internal/services/reminder"
	"first-max-bot/internal/services/schedule"
	"first-max-bot/internal/services/support"
	"first-max-bot/internal/services/user"
	redisstate "first-max-bot/internal/state/redis"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	logger := log.With().Str("component", "max_helper").Logger()

	api, err := maxbot.New(cfg.BotToken)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create max api client")
	}

	redisClient := redisclient.NewClient(&redisclient.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	stateRepo := redisstate.New(redisClient, redisstate.WithTTL(48*time.Hour))
	defer func() {
		if err := stateRepo.Close(); err != nil {
			logger.Error().Err(err).Msg("failed to close redis client")
		}
	}()

	if err := stateRepo.Ping(ctx); err != nil {
		logger.Fatal().Err(err).Msg("redis ping failed")
	}

	scheduleService := schedule.NewMock(cfg.MockScheduleLag)
	supportService := support.NewMock()
	userService := user.NewMock()
	deaneryService := deanery.NewMock()
	libraryService := library.NewMock()
	businessTripService := businesstrip.NewMock()
	newsService := news.NewMockService()
	moodleService := moodle.NewService()
	reminderService := reminder.NewMockService()

	// Инициализируем AI сервис (YandexGPT)
	var aiService ai.Service
	if cfg.YandexGPTAPIKey != "" && cfg.YandexGPTFolderID != "" {
		aiService = ai.NewYandexGPTService(cfg.YandexGPTAPIKey, cfg.YandexGPTFolderID)
		logger.Info().Msg("YandexGPT service initialized")
	} else {
		logger.Warn().Msg("YandexGPT API key or folder ID not set, AI service disabled")
		// Можно создать mock сервис или просто не регистрировать handler
		aiService = nil
	}

	router := botpkg.NewRouter()
	router.Register("/start", handlers.NewStartHandler(userService, logger.With().Str("handler", "start").Logger()))
	menuHandler := handlers.NewMenuHandler(userService)
	router.Register("/menu", menuHandler)
	router.Register("/help", menuHandler) // Используем тот же handler что и для /menu
	router.Register("/schedule", handlers.NewScheduleHandler(scheduleService, logger.With().Str("handler", "schedule").Logger()))
	router.Register("/contact", handlers.NewSupportHandler(supportService, logger.With().Str("handler", "support").Logger()))

	myTicketsHandler := handlers.NewMyTicketsHandler(supportService, logger.With().Str("handler", "mytickets").Logger())
	router.Register("/mytickets", myTicketsHandler)
	router.RegisterCallback("myticket:*", myTicketsHandler) // Регистрируем callback handler для обращений пользователя

	// Команды для абитуриентов
	router.Register("/admission", handlers.NewAdmissionHandler())
	router.Register("/programs", handlers.NewProgramsHandler())
	router.Register("/openday", handlers.NewOpenDayHandler())

	// Команды для студентов
	studentScheduleHandler := handlers.NewScheduleHandler(scheduleService, logger.With().Str("handler", "student_schedule").Logger())
	router.Register("/myschedule", studentScheduleHandler)

	deaneryHandler := handlers.NewDeaneryHandler(deaneryService, logger.With().Str("handler", "deanery").Logger())
	router.Register("/deanery", deaneryHandler)
	router.RegisterCallback("doc:*", deaneryHandler) // Регистрируем callback handler для документов

	libraryHandler := handlers.NewLibraryHandler(libraryService, userService, logger.With().Str("handler", "library").Logger())
	router.Register("/library", libraryHandler)
	router.RegisterCallback("book:*", libraryHandler) // Регистрируем callback handler для книг

	libraryManageHandler := handlers.NewLibraryManageHandler(libraryService, userService, logger.With().Str("handler", "library_manage").Logger())
	router.Register("/library_manage", libraryManageHandler)
	router.RegisterCallback("lib_manage:*", libraryManageHandler) // Регистрируем callback handler для управления библиотекой

	router.Register("/dormitory", handlers.NewDormitoryHandler())

	moodleHandler := handlers.NewMoodleHandler(moodleService, userService, logger.With().Str("handler", "moodle").Logger())
	router.Register("/moodle", moodleHandler)
	router.RegisterCallback("moodle:*", moodleHandler) // Регистрируем callback handler для Moodle

	reminderHandler := handlers.NewReminderHandler(reminderService, logger.With().Str("handler", "reminder").Logger())
	router.Register("/reminder", reminderHandler)
	router.RegisterCallback("reminder:*", reminderHandler) // Регистрируем callback handler для напоминаний

	// AI помощник (только если сервис инициализирован)
	if aiService != nil {
		askHandler := handlers.NewAskHandler(aiService, scheduleService, moodleService, userService, logger.With().Str("handler", "ask").Logger())
		router.Register("/ask", askHandler)
	}

	// TODO: /career, /projects, /events

	// Команды для сотрудников
	router.Register("/businesstrip", handlers.NewBusinessTripHandler(businessTripService, logger.With().Str("handler", "businesstrip").Logger()))
	router.Register("/vacation", handlers.NewVacationHandler())
	router.Register("/office", handlers.NewOfficeHandler())

	// Команды для руководителей
	router.Register("/dashboard", handlers.NewDashboardHandler())
	router.Register("/analytics", handlers.NewAnalyticsHandler())
	newsHandler := handlers.NewNewsHandler(newsService, logger.With().Str("handler", "news").Logger())
	router.Register("/news", newsHandler)

	sendNewsHandler := handlers.NewSendNewsHandler(newsService, userService, logger.With().Str("handler", "send_news").Logger())
	router.Register("/send_news", sendNewsHandler)

	ticketsHandler := handlers.NewTicketsHandler(supportService, userService, logger.With().Str("handler", "tickets").Logger())
	router.Register("/tickets", ticketsHandler)
	router.RegisterCallback("ticket:*", ticketsHandler) // Регистрируем callback handler для обращений

	documentsHandler := handlers.NewDocumentsHandler(deaneryService, userService, logger.With().Str("handler", "documents").Logger())
	router.Register("/documents", documentsHandler)
	router.RegisterCallback("doc_admin:*", documentsHandler) // Регистрируем callback handler для заявлений деканата

	// User registration handler
	userRegHandler := handlers.NewUserRegistrationHandler(userService, logger.With().Str("handler", "user_registration").Logger())
	router.Register("/register", userRegHandler)
	router.RegisterCallback("user_reg:*", userRegHandler)

	router.SetFallback(handlers.NewFallbackHandler())

	helperBot := botpkg.New(api, router, stateRepo, logger)

	// Запускаем фоновый процесс для проверки напоминаний
	go startReminderChecker(ctx, reminderService, api, logger.With().Str("component", "reminder_checker").Logger())

	logger.Info().Msg("max helper bot started")
	if err := helperBot.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error().Err(err).Msg("bot stopped with error")
	} else {
		logger.Info().Msg("bot stopped")
	}
}

// startReminderChecker запускает фоновый процесс для проверки и отправки напоминаний
func startReminderChecker(ctx context.Context, reminderService reminder.Service, api *maxbot.Api, logger zerolog.Logger) {
	ticker := time.NewTicker(1 * time.Minute) // Проверяем каждую минуту
	defer ticker.Stop()

	logger.Info().Msg("reminder checker started")

	// Первая проверка сразу при запуске
	checkAndSendReminders(ctx, reminderService, api, logger)

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("reminder checker stopped")
			return
		case <-ticker.C:
			checkAndSendReminders(ctx, reminderService, api, logger)
		}
	}
}

// checkAndSendReminders проверяет активные напоминания и отправляет те, которые должны быть отправлены
func checkAndSendReminders(ctx context.Context, reminderService reminder.Service, api *maxbot.Api, logger zerolog.Logger) {
	now := time.Now()

	reminders, err := reminderService.GetAllActiveReminders(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get active reminders")
		return
	}

	if len(reminders) == 0 {
		return
	}

	logger.Debug().Int("count", len(reminders)).Msg("checking reminders")

	for _, r := range reminders {
		if !r.DateTime.After(now) {
			if err := sendReminderToUser(ctx, api, r, logger); err != nil {
				logger.Error().Err(err).Str("reminder_id", r.ID).Str("user_id", r.UserID).Msg("failed to send reminder")
				continue
			}

			if err := reminderService.MarkReminderCompleted(ctx, r.ID); err != nil {
				logger.Error().Err(err).Str("reminder_id", r.ID).Msg("failed to mark reminder as completed")
			} else {
				logger.Info().Str("reminder_id", r.ID).Str("user_id", r.UserID).Str("text", r.Text).Msg("reminder sent and marked as completed")
			}
		}
	}
}

// sendReminderToUser отправляет напоминание пользователю
func sendReminderToUser(ctx context.Context, api *maxbot.Api, r reminder.Reminder, logger zerolog.Logger) error {
	// Парсим userID в int64
	userID, err := strconv.ParseInt(r.UserID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse user ID: %w", err)
	}

	// Формируем сообщение
	message := fmt.Sprintf("⏰ **Напоминание**\n\n%s", r.Text)

	// Создаем сообщение
	msg := maxbot.NewMessage()
	msg.SetUser(userID)
	msg.SetText(message)
	msg.SetFormat("markdown")

	// Отправляем сообщение
	_, err = api.Messages.Send(ctx, msg)
	if err != nil {
		a, _ := err.(schemes.Error)
		if a.Code == "" {
			return nil
		}
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
