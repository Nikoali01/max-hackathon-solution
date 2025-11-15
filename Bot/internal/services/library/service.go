package library

import (
	"context"
	"fmt"
	"time"
)

type Book struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Author     string     `json:"author"`
	ISBN       string     `json:"isbn"`
	Available  bool       `json:"available"`
	ReturnDate *time.Time `json:"return_date,omitempty"`
}

type UserBook struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	BookID      string     `json:"book_id"`
	Book        *Book      `json:"book,omitempty"`
	UserName    string     `json:"user_name"`    // Имя пользователя
	UserSurname string     `json:"user_surname"` // Фамилия пользователя
	Status      string     `json:"status"`       // "requested", "issued", "taken", "returned"
	BorrowedAt  time.Time  `json:"borrowed_at"`
	ReturnDate  time.Time  `json:"return_date"`
	Returned    bool       `json:"returned"`
	IssuedAt    *time.Time `json:"issued_at,omitempty"`
	TakenAt     *time.Time `json:"taken_at,omitempty"` // Когда книга была забрана
}

type Service interface {
	SearchBooks(ctx context.Context, query string) ([]Book, error)
	GetBookByID(ctx context.Context, bookID string) (*Book, error) // Получить книгу по ID
	BorrowBook(ctx context.Context, userID, userName, userSurname, bookID string) (*UserBook, error)
	GetUserBooks(ctx context.Context, userID string) ([]UserBook, error)
	ReturnBook(ctx context.Context, userID, bookID string) error
	GetAllRequests(ctx context.Context) ([]UserBook, error)                         // Для учителей - все запросы
	IssueBook(ctx context.Context, userID string, bookID string) (*UserBook, error) // Отметить как "выдано", возвращает UserBook для уведомления
	MarkBookTaken(ctx context.Context, userID string, bookID string) error          // Отметить как "забрано"
	MarkBookReturned(ctx context.Context, userID string, bookID string) error       // Отметить как "возвращена в библиотеку"
}

type mockService struct {
	books     map[string]*Book
	userBooks map[string]*UserBook
}

func NewMock() Service {
	service := &mockService{
		books:     make(map[string]*Book),
		userBooks: make(map[string]*UserBook),
	}

	// Добавляем мок-книги
	service.books["1"] = &Book{ID: "1", Title: "Введение в алгоритмы", Author: "Томас Кормен", ISBN: "978-5-8459-0857-4", Available: true}
	service.books["2"] = &Book{ID: "2", Title: "Чистый код", Author: "Роберт Мартин", ISBN: "978-5-4461-0772-1", Available: true}
	service.books["3"] = &Book{ID: "3", Title: "Архитектура компьютера", Author: "Эндрю Таненбаум", ISBN: "978-5-4461-1234-3", Available: true}
	service.books["4"] = &Book{ID: "4", Title: "Дизайн паттерны", Author: "Gang of Four", ISBN: "978-5-459-00401-2", Available: true}

	return service
}

func (s *mockService) SearchBooks(ctx context.Context, query string) ([]Book, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var result []Book
	// Простой поиск - возвращаем все доступные книги
	// Книга доступна, если она не занята (нет активных запросов/выдач/забрана)
	for _, book := range s.books {
		// Проверяем, есть ли активные запросы на эту книгу
		isBooked := false
		for _, ub := range s.userBooks {
			if ub.BookID == book.ID && !ub.Returned && (ub.Status == "requested" || ub.Status == "issued" || ub.Status == "taken") {
				// Книга занята, если она запрошена, выдана или забрана
				isBooked = true
				break
			}
		}

		// Книга доступна, если она не занята (requested, issued или taken)
		// Книги со статусом "returned" снова доступны
		if !isBooked {
			result = append(result, *book)
		}
	}
	return result, nil
}

func (s *mockService) GetBookByID(ctx context.Context, bookID string) (*Book, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	book, exists := s.books[bookID]
	if !exists {
		return nil, fmt.Errorf("book not found")
	}
	return book, nil
}

func (s *mockService) BorrowBook(ctx context.Context, userID, userName, userSurname, bookID string) (*UserBook, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	book, exists := s.books[bookID]
	if !exists || !book.Available {
		return nil, fmt.Errorf("книга недоступна")
	}

	now := time.Now()
	returnDate := now.Add(30 * 24 * time.Hour) // 30 дней

	userBook := &UserBook{
		ID:          fmt.Sprintf("UB-%d", now.Unix()),
		UserID:      userID,
		BookID:      bookID,
		Book:        book,
		UserName:    userName,
		UserSurname: userSurname,
		Status:      "requested",
		BorrowedAt:  now,
		ReturnDate:  returnDate,
		Returned:    false,
	}

	s.userBooks[userBook.ID] = userBook
	book.Available = false
	return userBook, nil
}

func (s *mockService) GetAllRequests(ctx context.Context) ([]UserBook, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var result []UserBook
	for _, ub := range s.userBooks {
		// Возвращаем только не возвращенные книги
		if !ub.Returned && (ub.Status == "requested" || ub.Status == "issued" || ub.Status == "taken") {
			result = append(result, *ub)
		}
	}
	return result, nil
}

func (s *mockService) IssueBook(ctx context.Context, userID string, bookID string) (*UserBook, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Находим запрос на книгу
	for _, ub := range s.userBooks {
		if ub.UserID == userID && ub.BookID == bookID && ub.Status == "requested" {
			now := time.Now()
			ub.Status = "issued"
			ub.IssuedAt = &now
			return ub, nil
		}
	}
	return nil, fmt.Errorf("request not found")
}

func (s *mockService) MarkBookTaken(ctx context.Context, userID string, bookID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Находим выданную книгу
	for _, ub := range s.userBooks {
		if ub.UserID == userID && ub.BookID == bookID && ub.Status == "issued" {
			now := time.Now()
			ub.Status = "taken"
			ub.TakenAt = &now
			return nil
		}
	}
	return fmt.Errorf("issued book not found")
}

func (s *mockService) MarkBookReturned(ctx context.Context, userID string, bookID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Находим забранную книг
	for _, ub := range s.userBooks {
		if ub.UserID == userID && ub.BookID == bookID && ub.Status == "taken" {
			ub.Status = "returned"
			ub.Returned = true
			s.books[bookID].Available = true
			return nil
		}
	}
	return fmt.Errorf("taken book not found")
}

func (s *mockService) GetUserBooks(ctx context.Context, userID string) ([]UserBook, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var result []UserBook
	for _, ub := range s.userBooks {
		// Показываем только книги, которые запрошены, выданы или забраны (но не возвращены)
		// Книги со статусом "returned" не показываются, так как они уже возвращены
		if ub.UserID == userID && !ub.Returned {
			result = append(result, *ub)
		}
	}
	return result, nil
}

func (s *mockService) ReturnBook(ctx context.Context, userID, bookID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, ub := range s.userBooks {
		if ub.UserID == userID && ub.BookID == bookID && !ub.Returned {
			ub.Returned = true
			ub.Status = "returned"
			// Книга автоматически станет доступной, так как SearchBooks проверяет активные запросы
			return nil
		}
	}
	return fmt.Errorf("книга не найдена")
}
