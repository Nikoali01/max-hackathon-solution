package schedule

import (
	"context"
	"time"
)

type Item struct {
	Time        time.Time
	Discipline  string
	Instructor  string
	Location    string
	Description string
}

type Service interface {
	GetSchedule(ctx context.Context, userID string) ([]Item, error)
}

type mockService struct {
	lag time.Duration
}

func NewMock(lag time.Duration) Service {
	return &mockService{
		lag: lag,
	}
}

func (m *mockService) GetSchedule(ctx context.Context, userID string) ([]Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(m.lag):
	}

	now := time.Now()
	return []Item{
		{
			Time:        now.Add(2 * time.Hour),
			Discipline:  "Мат. анализ",
			Instructor:  "доц. Светлана Иванова",
			Location:    "Корпус А, ауд. 302",
			Description: "Лекция. Возьмите тетрадь и калькулятор.",
		},
		{
			Time:        now.Add(5 * time.Hour),
			Discipline:  "Программирование",
			Instructor:  "проф. Алексей Петров",
			Location:    "Корпус Б, ауд. 115",
			Description: "Практика по Go. Подготовьте вопросы по goroutines.",
		},
	}, nil
}
