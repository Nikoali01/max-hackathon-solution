package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/library"
	"first-max-bot/internal/services/user"
)

// LibraryManageHandler обрабатывает команду /library_manage для учителей
type LibraryManageHandler struct {
	libraryService library.Service
	userService    user.Service
	logger         zerolog.Logger
}

func NewLibraryManageHandler(libraryService library.Service, userService user.Service, logger zerolog.Logger) *LibraryManageHandler {
	return &LibraryManageHandler{
		libraryService: libraryService,
		userService:    userService,
		logger:         logger,
	}
}

func (h *LibraryManageHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Проверяем, что пользователь - сотрудник (учитель)
	userID := req.UserID()
	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil || u == nil || (u.Role != user.RoleEmployee && u.Role != user.RoleManager) {
		return responder.SendText(ctx, req.Recipient(), "❌ Эта команда доступна только сотрудникам и руководителям.")
	}

	// Если это callback для управления книгами
	if strings.HasPrefix(req.Args, "lib_manage:") {
		return h.handleManageCallback(ctx, req, responder)
	}

	// Получаем все запросы на книги
	requests, err := h.libraryService.GetAllRequests(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get book requests")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении запросов на книги")
	}

	h.logger.Debug().Int("total_requests", len(requests)).Msg("got requests from service")

	// GetAllRequests уже возвращает только активные запросы (requested, issued, taken)
	activeRequests := requests

	var message strings.Builder
	message.WriteString("📚 Управление библиотекой\n\n")

	// Группируем по статусу
	requested := []library.UserBook{}
	issued := []library.UserBook{}
	taken := []library.UserBook{}

	for _, req := range activeRequests {
		h.logger.Debug().Str("book_id", req.BookID).Str("user_id", req.UserID).Str("status", req.Status).Msg("processing request")
		switch req.Status {
		case "requested":
			requested = append(requested, req)
		case "issued":
			issued = append(issued, req)
		case "taken":
			taken = append(taken, req)
		}
	}

	h.logger.Debug().Int("requested", len(requested)).Int("issued", len(issued)).Int("taken", len(taken)).Msg("grouped by status")

	totalActive := len(requested) + len(issued) + len(taken)
	if totalActive == 0 {
		message.WriteString("✅ Нет активных запросов на книги.")
		return responder.SendText(ctx, req.Recipient(), message.String())
	}

	message.WriteString(fmt.Sprintf("Активных запросов: %d\n\n", totalActive))

	keyboard := responder.NewKeyboardBuilder()

	// Показываем запрошенные книги
	if len(requested) > 0 {
		message.WriteString("⏳ Запрошенные книги:\n")
		for i, req := range requested {
			if i >= 10 {
				break
			}
			book, _ := h.libraryService.GetBookByID(ctx, req.BookID)
			bookTitle := "Неизвестная книга"
			if book != nil {
				bookTitle = book.Title
			}
			userName := fmt.Sprintf("%s %s", req.UserName, req.UserSurname)
			if userName == "Имя Фамилия" {
				// Пытаемся получить реальное имя пользователя
				u, _ := h.userService.GetUserByID(ctx, req.UserID)
				if u != nil {
					userName = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
				} else {
					userName = req.UserID
				}
			}
			message.WriteString(fmt.Sprintf("• %s — %s\n", bookTitle, userName))

			row := keyboard.AddRow()
			buttonText := fmt.Sprintf("✅ Выдано: %s", bookTitle)
			if len(buttonText) > 40 {
				buttonText = buttonText[:37] + "..."
			}
			row.AddCallback(buttonText, schemes.POSITIVE, fmt.Sprintf("lib_manage:issue:%s:%s", req.UserID, req.BookID))
		}
		message.WriteString("\n")
	}

	// Показываем выданные книги (ожидают, что заберут)
	if len(issued) > 0 {
		message.WriteString("📦 Выданные (ожидают получения):\n")
		for i, req := range issued {
			if i >= 10 {
				break
			}
			book, _ := h.libraryService.GetBookByID(ctx, req.BookID)
			bookTitle := "Неизвестная книга"
			if book != nil {
				bookTitle = book.Title
			}
			userName := fmt.Sprintf("%s %s", req.UserName, req.UserSurname)
			if userName == " " || (req.UserName == "" && req.UserSurname == "") {
				u, _ := h.userService.GetUserByID(ctx, req.UserID)
				if u != nil {
					userName = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
				} else {
					userName = req.UserID
				}
			}
			message.WriteString(fmt.Sprintf("• %s — %s\n", bookTitle, userName))

			row := keyboard.AddRow()
			buttonText := fmt.Sprintf("✅ Забрано: %s", bookTitle)
			if len(buttonText) > 40 {
				buttonText = buttonText[:37] + "..."
			}
			row.AddCallback(buttonText, schemes.POSITIVE, fmt.Sprintf("lib_manage:taken:%s:%s", req.UserID, req.BookID))
		}
		message.WriteString("\n")
	}

	// Показываем забранные книги
	if len(taken) > 0 {
		message.WriteString("📖 Забранные книги:\n")
		for i, req := range taken {
			if i >= 10 {
				break
			}
			book, _ := h.libraryService.GetBookByID(ctx, req.BookID)
			bookTitle := "Неизвестная книга"
			if book != nil {
				bookTitle = book.Title
			}
			userName := fmt.Sprintf("%s %s", req.UserName, req.UserSurname)
			if userName == " " || (req.UserName == "" && req.UserSurname == "") {
				u, _ := h.userService.GetUserByID(ctx, req.UserID)
				if u != nil {
					userName = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
				} else {
					userName = req.UserID
				}
			}
			takenTime := ""
			if req.TakenAt != nil {
				takenTime = req.TakenAt.Format("02.01.2006 15:04")
			}
			message.WriteString(fmt.Sprintf("• %s — %s (забрано: %s)\n", bookTitle, userName, takenTime))

			row := keyboard.AddRow()
			buttonText := fmt.Sprintf("📚 Вернулась: %s", bookTitle)
			if len(buttonText) > 40 {
				buttonText = buttonText[:37] + "..."
			}
			row.AddCallback(buttonText, schemes.POSITIVE, fmt.Sprintf("lib_manage:returned:%s:%s", req.UserID, req.BookID))
		}
		message.WriteString("\n")
	}

	return responder.SendTextWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
}

func (h *LibraryManageHandler) handleManageCallback(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	payload := req.Args

	callbackID := ""
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok && cid != "" {
			callbackID = cid
			responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
		}
	}

	if strings.HasPrefix(payload, "lib_manage:issue:") {
		// Формат: lib_manage:issue:userID:bookID
		parts := strings.Split(payload, ":")
		if len(parts) != 4 {
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка: неверный формат запроса")
		}
		userID := parts[2]
		bookID := parts[3]

		userBook, err := h.libraryService.IssueBook(ctx, userID, bookID)
		if err != nil {
			h.logger.Error().Err(err).Str("user_id", userID).Str("book_id", bookID).Msg("failed to issue book")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при выдаче книги")
		}

		// Получаем информацию о книге и пользователе
		book, _ := h.libraryService.GetBookByID(ctx, bookID)
		bookTitle := "книга"
		if book != nil {
			bookTitle = book.Title
		}
		u, _ := h.userService.GetUserByID(ctx, userID)
		userName := userID
		if u != nil {
			userName = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
		}

		// Отправляем уведомление пользователю
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err == nil {
			userRecipient := schemes.Recipient{
				UserId:   userIDInt,
				ChatType: schemes.DIALOG,
			}

			notification := fmt.Sprintf("✅ Книга \"%s\" готова к выдаче!\n\n", bookTitle)
			if userBook != nil && userBook.ReturnDate != (time.Time{}) {
				notification += fmt.Sprintf("Срок возврата: %s\n\n", userBook.ReturnDate.Format("02.01.2006"))
			}
			notification += "Можешь забрать книгу в библиотеке."

			if err := responder.SendText(ctx, userRecipient, notification); err != nil {
				h.logger.Warn().Err(err).Str("user_id", userID).Msg("failed to send notification to user")
			} else {
				h.logger.Info().Str("book_id", bookID).Str("user_id", userID).Msg("user notified about book ready")
			}
		}

		message := fmt.Sprintf("✅ Книга \"%s\" отмечена как готовая к выдаче.\n\n", bookTitle)
		message += fmt.Sprintf("Пользователь %s получил уведомление.\n\n", userName)
		message += "Используй /library_manage чтобы посмотреть все запросы."

		return responder.SendText(ctx, req.Recipient(), message)
	}

	if strings.HasPrefix(payload, "lib_manage:taken:") {
		// Формат: lib_manage:taken:userID:bookID
		parts := strings.Split(payload, ":")
		if len(parts) != 4 {
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка: неверный формат запроса")
		}
		userID := parts[2]
		bookID := parts[3]

		err := h.libraryService.MarkBookTaken(ctx, userID, bookID)
		if err != nil {
			h.logger.Error().Err(err).Str("user_id", userID).Str("book_id", bookID).Msg("failed to mark book as taken")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при отметке книги как забранной")
		}

		// Получаем информацию о книге и пользователе
		book, _ := h.libraryService.GetBookByID(ctx, bookID)
		bookTitle := "книга"
		if book != nil {
			bookTitle = book.Title
		}
		u, _ := h.userService.GetUserByID(ctx, userID)
		userName := userID
		if u != nil {
			userName = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
		}

		message := fmt.Sprintf("✅ Книга \"%s\" отмечена как забранная пользователем %s.\n\n", bookTitle, userName)
		message += "Используй /library_manage чтобы посмотреть все запросы."

		return responder.SendText(ctx, req.Recipient(), message)
	}

	if strings.HasPrefix(payload, "lib_manage:returned:") {
		// Формат: lib_manage:returned:userID:bookID
		parts := strings.Split(payload, ":")
		if len(parts) != 4 {
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка: неверный формат запроса")
		}
		userID := parts[2]
		bookID := parts[3]

		err := h.libraryService.MarkBookReturned(ctx, userID, bookID)
		if err != nil {
			h.logger.Error().Err(err).Str("user_id", userID).Str("book_id", bookID).Msg("failed to mark book as returned")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при отметке книги как возвращенной")
		}

		// Получаем информацию о книге и пользователе
		book, _ := h.libraryService.GetBookByID(ctx, bookID)
		bookTitle := "книга"
		if book != nil {
			bookTitle = book.Title
		}
		u, _ := h.userService.GetUserByID(ctx, userID)
		userName := userID
		if u != nil {
			userName = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
		}

		message := fmt.Sprintf("✅ Книга \"%s\" отмечена как возвращенная в библиотеку пользователем %s.\n\n", bookTitle, userName)
		message += "Книга снова доступна для выдачи.\n\n"
		message += "Используй /library_manage чтобы посмотреть все запросы."

		return responder.SendText(ctx, req.Recipient(), message)
	}

	return nil
}
