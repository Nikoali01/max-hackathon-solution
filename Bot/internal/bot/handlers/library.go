package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/library"
	"first-max-bot/internal/services/user"
)

// LibraryHandler обрабатывает команду /library
type LibraryHandler struct {
	libraryService library.Service
	userService    user.Service
	logger         zerolog.Logger
}

func NewLibraryHandler(libraryService library.Service, userService user.Service, logger zerolog.Logger) *LibraryHandler {
	return &LibraryHandler{
		libraryService: libraryService,
		userService:    userService,
		logger:         logger,
	}
}

func (h *LibraryHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Если это callback для заказа книги
	if strings.HasPrefix(req.Args, "book:") {
		return h.handleBookCallback(ctx, req, responder)
	}

	userID := req.UserID()
	if userID == "" {
		return responder.SendText(ctx, req.Recipient(), "Не удалось определить пользователя")
	}

	// Получаем книги пользователя
	userBooks, err := h.libraryService.GetUserBooks(ctx, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get user books")
	}

	// Получаем доступные книги
	availableBooks, err := h.libraryService.SearchBooks(ctx, "")
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to search books")
	}

	var message strings.Builder
	message.WriteString("📚 Библиотека\n\n")

	if len(userBooks) > 0 {
		message.WriteString("📖 Твои книги:\n")
		for _, ub := range userBooks {
			statusLabel := ""
			switch ub.Status {
			case "requested":
				statusLabel = "⏳ Запрошена"
			case "issued":
				statusLabel = "✅ Готова к выдаче"
			case "taken":
				statusLabel = "📖 У тебя"
			default:
				statusLabel = "📄 " + ub.Status
			}
			message.WriteString(fmt.Sprintf("• %s (%s) — %s\n", ub.Book.Title, ub.Book.Author, statusLabel))
			if ub.ReturnDate != (time.Time{}) {
				message.WriteString(fmt.Sprintf("  Срок возврата: %s\n", ub.ReturnDate.Format("02.01.2006")))
			}
		}
		message.WriteString("\n")
	}

	message.WriteString("Доступные книги для заказа:\n")

	keyboard := responder.NewKeyboardBuilder()
	
	// Показываем первые 4 доступные книги
	for i, book := range availableBooks {
		if i >= 4 {
			break
		}
		row := keyboard.AddRow()
		row.AddCallback(fmt.Sprintf("📖 %s", book.Title), schemes.POSITIVE, fmt.Sprintf("book:borrow:%s", book.ID))
	}

	if len(availableBooks) == 0 {
		message.WriteString("Нет доступных книг в данный момент.")
	}

	return responder.SendTextWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
}

func (h *LibraryHandler) handleBookCallback(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	payload := req.Args
	userID := req.UserID()

	callbackID := ""
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok && cid != "" {
			callbackID = cid
			responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
		}
	}

	if strings.HasPrefix(payload, "book:borrow:") {
		bookID := strings.TrimPrefix(payload, "book:borrow:")
		
		// Получаем информацию о пользователе для сохранения имени и фамилии
		u, err := h.userService.GetUserByID(ctx, userID)
		userName := ""
		userSurname := ""
		if err == nil && u != nil {
			userName = u.FirstName
			userSurname = u.LastName
		}
		
		userBook, err := h.libraryService.BorrowBook(ctx, userID, userName, userSurname, bookID)
		if err != nil {
			return responder.SendText(ctx, req.Recipient(), fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		}

		message := fmt.Sprintf("✅ Книга заказана!\n\n")
		message += fmt.Sprintf("📖 %s\n", userBook.Book.Title)
		message += fmt.Sprintf("Автор: %s\n", userBook.Book.Author)
		message += fmt.Sprintf("Срок возврата: %s\n\n", userBook.ReturnDate.Format("02.01.2006"))
		message += "Книга будет готова к выдаче в течение 1-2 рабочих дней."

		return responder.SendText(ctx, req.Recipient(), message)
	}

	return nil
}

