package reminder

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Reminder представляет напоминание
type Reminder struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	DateTime  time.Time `json:"date_time"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // "active", "completed", "cancelled"
}

// Service определяет интерфейс для работы с напоминаниями
type Service interface {
	CreateReminder(ctx context.Context, userID, text string, dateTime time.Time) (*Reminder, error)
	GetUserReminders(ctx context.Context, userID string) ([]Reminder, error)
	GetActiveReminders(ctx context.Context, userID string) ([]Reminder, error)
	GetAllActiveReminders(ctx context.Context) ([]Reminder, error) // Все активные напоминания всех пользователей
	DeleteReminder(ctx context.Context, reminderID string) error
	GetReminderByID(ctx context.Context, reminderID string) (*Reminder, error)
	MarkReminderCompleted(ctx context.Context, reminderID string) error // Пометить как выполненное
}

type mockService struct {
	reminders map[string]*Reminder
	mu        sync.RWMutex // Для thread-safety
}

func NewMockService() Service {
	return &mockService{
		reminders: make(map[string]*Reminder),
		mu:        sync.RWMutex{},
	}
}

func (s *mockService) CreateReminder(ctx context.Context, userID, text string, dateTime time.Time) (*Reminder, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	now := time.Now()
	reminderID := fmt.Sprintf("REM-%d", now.UnixNano())
	
	reminder := &Reminder{
		ID:        reminderID,
		UserID:    userID,
		Text:      text,
		DateTime:  dateTime,
		CreatedAt: now,
		Status:    "active",
	}

	s.mu.Lock()
	s.reminders[reminderID] = reminder
	s.mu.Unlock()
	return reminder, nil
}

func (s *mockService) GetUserReminders(ctx context.Context, userID string) ([]Reminder, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	var userReminders []Reminder
	for _, r := range s.reminders {
		if r.UserID == userID {
			userReminders = append(userReminders, *r)
		}
	}
	s.mu.RUnlock()

	// Сортируем по дате (ближайшие первыми)
	sort.Slice(userReminders, func(i, j int) bool {
		return userReminders[i].DateTime.Before(userReminders[j].DateTime)
	})

	return userReminders, nil
}

func (s *mockService) GetActiveReminders(ctx context.Context, userID string) ([]Reminder, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	var activeReminders []Reminder
	for _, r := range s.reminders {
		if r.UserID == userID && r.Status == "active" {
			activeReminders = append(activeReminders, *r)
		}
	}
	s.mu.RUnlock()

	// Сортируем по дате (ближайшие первыми)
	sort.Slice(activeReminders, func(i, j int) bool {
		return activeReminders[i].DateTime.Before(activeReminders[j].DateTime)
	})

	return activeReminders, nil
}

func (s *mockService) DeleteReminder(ctx context.Context, reminderID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	delete(s.reminders, reminderID)
	s.mu.Unlock()
	return nil
}

func (s *mockService) GetReminderByID(ctx context.Context, reminderID string) (*Reminder, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	reminder, ok := s.reminders[reminderID]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("reminder not found")
	}

	// Возвращаем копию, чтобы избежать race conditions
	reminderCopy := *reminder
	return &reminderCopy, nil
}

func (s *mockService) GetAllActiveReminders(ctx context.Context) ([]Reminder, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	var activeReminders []Reminder
	for _, r := range s.reminders {
		if r.Status == "active" {
			activeReminders = append(activeReminders, *r)
		}
	}
	s.mu.RUnlock()

	// Сортируем по дате (ближайшие первыми)
	sort.Slice(activeReminders, func(i, j int) bool {
		return activeReminders[i].DateTime.Before(activeReminders[j].DateTime)
	})

	return activeReminders, nil
}

func (s *mockService) MarkReminderCompleted(ctx context.Context, reminderID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	reminder, ok := s.reminders[reminderID]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("reminder not found")
	}

	reminder.Status = "completed"
	s.mu.Unlock()
	return nil
}

