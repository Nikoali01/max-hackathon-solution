package support

import (
	"context"
	"fmt"
	"time"
)

type Ticket struct {
	ID              string
	UserID          string
	Department      string
	Subject         string
	Message         string
	Response        string   // Ответ руководителя на обращение
	ResponseBy      string   // ID администратора, который ответил
	UserReply       string   // Ответ пользователя на ответ руководителя
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Status          string // "received", "in_progress", "answered", "resolved", "closed"
}

type Service interface {
	CreateTicket(ctx context.Context, userID, subject, message string) (Ticket, error)
	GetTicket(ctx context.Context, ticketID string) (*Ticket, error)
	GetAllTickets(ctx context.Context) ([]Ticket, error)
	GetUserTickets(ctx context.Context, userID string) ([]Ticket, error)
	UpdateTicketStatus(ctx context.Context, ticketID, status string) error
	AddResponse(ctx context.Context, ticketID, response, responseBy string) error
	AddUserReply(ctx context.Context, ticketID, reply string) error
}

type mockService struct {
	tickets map[string]*Ticket
}

func NewMock() Service {
	return &mockService{
		tickets: make(map[string]*Ticket),
	}
}

func (s *mockService) CreateTicket(ctx context.Context, userID, subject, message string) (Ticket, error) {
	select {
	case <-ctx.Done():
		return Ticket{}, ctx.Err()
	default:
	}

	now := time.Now()
	ticket := &Ticket{
		ID:         fmt.Sprintf("DOE-%d", now.Unix()),
		UserID:     userID,
		Department: "Department of Education",
		Subject:    subject,
		Message:    message,
		CreatedAt:  now,
		UpdatedAt:  now,
		Status:     "received",
	}

	s.tickets[ticket.ID] = ticket
	return *ticket, nil
}

func (s *mockService) GetTicket(ctx context.Context, ticketID string) (*Ticket, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return nil, nil
	}
	return ticket, nil
}

func (s *mockService) GetAllTickets(ctx context.Context) ([]Ticket, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var result []Ticket
	for _, ticket := range s.tickets {
		result = append(result, *ticket)
	}
	return result, nil
}

func (s *mockService) GetUserTickets(ctx context.Context, userID string) ([]Ticket, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var result []Ticket
	for _, ticket := range s.tickets {
		if ticket.UserID == userID {
			result = append(result, *ticket)
		}
	}
	return result, nil
}

func (s *mockService) UpdateTicketStatus(ctx context.Context, ticketID, status string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return fmt.Errorf("ticket not found")
	}

	ticket.Status = status
	ticket.UpdatedAt = time.Now()
	return nil
}

func (s *mockService) AddResponse(ctx context.Context, ticketID, response, responseBy string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return fmt.Errorf("ticket not found")
	}

	ticket.Response = response
	ticket.ResponseBy = responseBy // Сохраняем ID администратора, который ответил
	ticket.Status = "answered"     // Статус "answered" - есть ответ, но тикет еще не закрыт
	ticket.UpdatedAt = time.Now()
	return nil
}

func (s *mockService) AddUserReply(ctx context.Context, ticketID, reply string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return fmt.Errorf("ticket not found")
	}

	// Если уже был ответ пользователя, добавляем новый ответ через перенос строки
	if ticket.UserReply != "" {
		ticket.UserReply = ticket.UserReply + "\n\n---\n\n" + reply
	} else {
		ticket.UserReply = reply
	}
	ticket.Status = "in_progress" // Возвращаем статус "in_progress" - ждем ответа руководителя
	ticket.UpdatedAt = time.Now()
	return nil
}
