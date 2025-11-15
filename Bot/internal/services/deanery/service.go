package deanery

import (
	"context"
	"fmt"
	"time"
)

type DocumentType string

const (
	DocumentTypeCertificate DocumentType = "certificate" // Справка
	DocumentTypePayment     DocumentType = "payment"     // Оплата обучения
	DocumentTypeTransfer    DocumentType = "transfer"    // Перевод
	DocumentTypeAcademicLeave DocumentType = "academic_leave" // Академический отпуск
)

type Document struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Type        DocumentType `json:"type"`
	Status      string       `json:"status"` // "pending", "approved", "rejected", "completed"
	Description string       `json:"description"`
	Response    string       `json:"response"`     // Ответ администратора (текст)
	ResponseFile string      `json:"response_file"` // Файл в ответе (URL или ID файла)
	ResponseBy  string       `json:"response_by"`   // ID администратора, который ответил
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Service interface {
	CreateDocument(ctx context.Context, userID string, docType DocumentType, description string) (*Document, error)
	GetUserDocuments(ctx context.Context, userID string) ([]Document, error)
	GetDocument(ctx context.Context, documentID string) (*Document, error)
	GetAllDocuments(ctx context.Context) ([]Document, error) // Для администраторов
	AddDocumentResponse(ctx context.Context, documentID, response, responseFile, responseBy string) error
}

type mockService struct {
	documents map[string]*Document
}

func NewMock() Service {
	return &mockService{
		documents: make(map[string]*Document),
	}
}

func (s *mockService) CreateDocument(ctx context.Context, userID string, docType DocumentType, description string) (*Document, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	now := time.Now()
	doc := &Document{
		ID:          fmt.Sprintf("DOC-%d", now.Unix()),
		UserID:      userID,
		Type:        docType,
		Status:      "pending",
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.documents[doc.ID] = doc
	return doc, nil
}

func (s *mockService) GetUserDocuments(ctx context.Context, userID string) ([]Document, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var result []Document
	for _, doc := range s.documents {
		if doc.UserID == userID {
			result = append(result, *doc)
		}
	}
	return result, nil
}

func (s *mockService) GetDocument(ctx context.Context, documentID string) (*Document, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	doc, exists := s.documents[documentID]
	if !exists {
		return nil, nil
	}
	return doc, nil
}

func (s *mockService) GetAllDocuments(ctx context.Context) ([]Document, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var result []Document
	for _, doc := range s.documents {
		result = append(result, *doc)
	}
	return result, nil
}

func (s *mockService) AddDocumentResponse(ctx context.Context, documentID, response, responseFile, responseBy string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	doc, exists := s.documents[documentID]
	if !exists {
		return fmt.Errorf("document not found")
	}

	doc.Response = response
	doc.ResponseFile = responseFile
	doc.ResponseBy = responseBy
	doc.Status = "completed"
	doc.UpdatedAt = time.Now()
	return nil
}

