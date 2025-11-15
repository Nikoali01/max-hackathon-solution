package user

import (
	"context"
	"fmt"
	"time"
)

type Role string

const (
	RoleApplicant Role = "applicant" // Абитуриент
	RoleStudent   Role = "student"   // Студент
	RoleEmployee  Role = "employee"  // Сотрудник
	RoleManager   Role = "manager"   // Руководитель
)

type User struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"` // ID пользователя в мессенджере
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Age         int       `json:"age"`
	Gender      string    `json:"gender"` // "male", "female"
	Email       string    `json:"email"`
	Role        Role      `json:"role"`
	MoodleToken string    `json:"moodle_token,omitempty"` // Токен для Moodle API
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Service interface {
	GetUserByID(ctx context.Context, userID string) (*User, error)
	CreateUser(ctx context.Context, user User) (*User, error)
	UpdateUser(ctx context.Context, userID string, user User) (*User, error)
	GetUserRole(ctx context.Context, userID string) (Role, error)
	GetAllUsers(ctx context.Context) ([]User, error)                       // Получить всех пользователей
	SetMoodleToken(ctx context.Context, userID string, token string) error // Установить токен Moodle
}

type mockService struct {
	users map[string]*User
}

func NewMock() Service {
	//newUser1 := User{
	//	UserID:    "4915056",
	//	FirstName: "Андрей",
	//	LastName:  "Алексеев",
	//	Age:       21,
	//	Gender:    "male",
	//	Email:     "jandr21@ya.ru",
	//	Role:      RoleStudent,
	//}
	//newUser2 := User{
	//	UserID:    "90721362",
	//	FirstName: "Николай",
	//	LastName:  "Кузмин",
	//	Age:       21,
	//	Gender:    "male",
	//	Email:     "englia228@gmail.com",
	//	Role:      RoleManager,
	//}
	users := make(map[string]*User)
	//users[newUser1.UserID] = &newUser1
	//users[newUser2.UserID] = &newUser2
	return &mockService{
		users: users,
	}
}

func (s *mockService) GetUserByID(ctx context.Context, userID string) (*User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	user, exists := s.users[userID]
	if !exists {
		return nil, nil // Пользователь не найден
	}
	return user, nil
}

func (s *mockService) CreateUser(ctx context.Context, user User) (*User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	now := time.Now()
	user.ID = user.UserID // Используем userID как ID
	user.CreatedAt = now
	user.UpdatedAt = now

	if user.Role == "" {
		user.Role = RoleApplicant
	}

	s.users[user.UserID] = &user
	return &user, nil
}

func (s *mockService) UpdateUser(ctx context.Context, userID string, user User) (*User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	existing, exists := s.users[userID]
	if !exists {
		return nil, nil
	}

	// Обновляем поля
	if user.FirstName != "" {
		existing.FirstName = user.FirstName
	}
	if user.LastName != "" {
		existing.LastName = user.LastName
	}
	if user.Age != 0 {
		existing.Age = user.Age
	}
	if user.Gender != "" {
		existing.Gender = user.Gender
	}
	if user.Email != "" {
		existing.Email = user.Email
	}
	if user.Role != "" {
		existing.Role = user.Role
	}
	existing.UpdatedAt = time.Now()

	return existing, nil
}

func (s *mockService) GetUserRole(ctx context.Context, userID string) (Role, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return RoleApplicant, nil // По умолчанию абитуриент
	}
	return user.Role, nil
}

func (s *mockService) GetAllUsers(ctx context.Context) ([]User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var result []User
	for _, u := range s.users {
		result = append(result, *u)
	}
	return result, nil
}

func (s *mockService) SetMoodleToken(ctx context.Context, userID string, token string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	user, exists := s.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.MoodleToken = token
	user.UpdatedAt = time.Now()
	return nil
}
