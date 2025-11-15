package news

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// News представляет новость
type News struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"` // Markdown формат
	AuthorID  string    `json:"author_id"`
	Author    string    `json:"author"` // Имя автора
	CreatedAt time.Time `json:"created_at"`
}

type Service interface {
	GetLatestNews(ctx context.Context, count int) ([]News, error)
	CreateNews(ctx context.Context, title, content, authorID, author string) (*News, error)
}

type mockService struct {
	news []*News
}

func NewMockService() Service {
	return &mockService{
		news: []*News{
			{
				ID:        "news-1",
				Title:     "Добро пожаловать в университет!",
				Content:   "**Добро пожаловать!**\n\nМы рады приветствовать всех новых студентов в нашем университете. Учебный год начинается с больших возможностей и новых вызовов.\n\n*Удачи в учебе!*",
				AuthorID:  "admin-1",
				Author:    "Администрация",
				CreatedAt: time.Now().Add(-5 * 24 * time.Hour),
			},
			{
				ID:        "news-2",
				Title:     "Обновление расписания",
				Content:   "**Важное обновление**\n\nРасписание занятий на следующую неделю было обновлено. Пожалуйста, проверьте актуальное расписание в разделе /schedule.\n\n*С уважением, Деканат*",
				AuthorID:  "admin-2",
				Author:    "Деканат",
				CreatedAt: time.Now().Add(-3 * 24 * time.Hour),
			},
			{
				ID:        "news-3",
				Title:     "День открытых дверей",
				Content:   "**Приглашаем на День открытых дверей!**\n\nПриглашаем всех желающих посетить наш университет 15 декабря. Вы сможете познакомиться с факультетами, преподавателями и студентами.\n\n*Регистрация: /openday*",
				AuthorID:  "admin-1",
				Author:    "Администрация",
				CreatedAt: time.Now().Add(-1 * 24 * time.Hour),
			},
		},
	}
}

func (s *mockService) GetLatestNews(ctx context.Context, count int) ([]News, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Сортируем по дате создания (новые первыми)
	newsCopy := make([]*News, len(s.news))
	copy(newsCopy, s.news)
	
	sort.Slice(newsCopy, func(i, j int) bool {
		return newsCopy[i].CreatedAt.After(newsCopy[j].CreatedAt)
	})

	// Берем последние count новостей
	result := []News{}
	for i := 0; i < count && i < len(newsCopy); i++ {
		result = append(result, *newsCopy[i])
	}

	return result, nil
}

func (s *mockService) CreateNews(ctx context.Context, title, content, authorID, author string) (*News, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	news := &News{
		ID:        fmt.Sprintf("news-%d", time.Now().Unix()),
		Title:     title,
		Content:   content,
		AuthorID:  authorID,
		Author:    author,
		CreatedAt: time.Now(),
	}

	s.news = append(s.news, news)
	return news, nil
}

