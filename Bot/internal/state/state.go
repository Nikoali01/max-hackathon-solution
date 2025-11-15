package state

import (
	"context"
	"time"
)

type UserState struct {
	LastCommand string    `json:"last_command"`
	LastUpdated time.Time `json:"last_updated"`

	// User registration state (новая регистрация пользователя)
	UserRegistrationStep string            `json:"user_registration_step,omitempty"` // "first_name", "last_name", "age", "gender", "email", "completed"
	UserRegistrationData map[string]string `json:"user_registration_data,omitempty"` // хранит данные регистрации
}

type Repository interface {
	GetUserState(ctx context.Context, userID string) (*UserState, error)
	SaveUserState(ctx context.Context, userID string, st UserState) error
	Ping(ctx context.Context) error
	Close() error
}
