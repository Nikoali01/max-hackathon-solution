package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/deanery"
)

// DeaneryHandler обрабатывает команду /deanery для студентов
type DeaneryHandler struct {
	deaneryService deanery.Service
	logger         zerolog.Logger
}

func NewDeaneryHandler(deaneryService deanery.Service, logger zerolog.Logger) *DeaneryHandler {
	return &DeaneryHandler{
		deaneryService: deaneryService,
		logger:         logger,
	}
}

func (h *DeaneryHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Если это callback для создания документа
	if strings.HasPrefix(req.Args, "doc:") {
		return h.handleDocumentCallback(ctx, req, responder)
	}

	userID := req.UserID()
	if userID == "" {
		return responder.SendText(ctx, req.Recipient(), "Не удалось определить пользователя")
	}

	// Получаем документы пользователя
	documents, err := h.deaneryService.GetUserDocuments(ctx, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get user documents")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении документов")
	}

	var message strings.Builder
	message.WriteString("🏛️ Деканат\n\n")
	message.WriteString("Доступные услуги:\n\n")

	keyboard := responder.NewKeyboardBuilder()
	
	row1 := keyboard.AddRow()
	row1.AddCallback("📄 Справка", schemes.POSITIVE, "doc:certificate")
	row1.AddCallback("💳 Оплата обучения", schemes.POSITIVE, "doc:payment")

	row2 := keyboard.AddRow()
	row2.AddCallback("🔄 Перевод", schemes.DEFAULT, "doc:transfer")
	row2.AddCallback("📋 Академический отпуск", schemes.DEFAULT, "doc:academic_leave")

	if len(documents) > 0 {
		message.WriteString("\n📋 Твои заявления:\n")
		for _, doc := range documents {
			statusEmoji := h.getStatusEmoji(doc.Status)
			message.WriteString(fmt.Sprintf("%s %s #%s — %s\n", statusEmoji, h.getDocumentTypeLabel(doc.Type), doc.ID, h.getStatusLabel(doc.Status)))
			
			// Если есть ответ, показываем его
			if doc.Response != "" {
				message.WriteString(fmt.Sprintf("   Ответ: %s\n", doc.Response))
			}
		}
	}

	return responder.SendTextWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
}

func (h *DeaneryHandler) handleDocumentCallback(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	payload := req.Args
	userID := req.UserID()

	callbackID := ""
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok && cid != "" {
			callbackID = cid
			responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
		}
	}

	docType := strings.TrimPrefix(payload, "doc:")
	var docTypeEnum deanery.DocumentType
	var description string

	switch docType {
	case "certificate":
		docTypeEnum = deanery.DocumentTypeCertificate
		description = "Запрос на получение справки"
	case "payment":
		docTypeEnum = deanery.DocumentTypePayment
		description = "Запрос на оплату обучения"
	case "transfer":
		docTypeEnum = deanery.DocumentTypeTransfer
		description = "Заявление на перевод"
	case "academic_leave":
		docTypeEnum = deanery.DocumentTypeAcademicLeave
		description = "Заявление на академический отпуск"
	default:
		return responder.SendText(ctx, req.Recipient(), "❌ Неизвестный тип документа")
	}

	doc, err := h.deaneryService.CreateDocument(ctx, userID, docTypeEnum, description)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create document")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при создании заявления")
	}

	message := fmt.Sprintf("✅ Заявление создано!\n\n")
	message += fmt.Sprintf("Тип: %s\n", h.getDocumentTypeLabel(doc.Type))
	message += fmt.Sprintf("Номер: %s\n", doc.ID)
	message += fmt.Sprintf("Статус: %s\n\n", doc.Status)
	message += "Твоё заявление будет рассмотрено в ближайшее время."

	return responder.SendText(ctx, req.Recipient(), message)
}

func (h *DeaneryHandler) getDocumentTypeLabel(docType deanery.DocumentType) string {
	switch docType {
	case deanery.DocumentTypeCertificate:
		return "Справка"
	case deanery.DocumentTypePayment:
		return "Оплата обучения"
	case deanery.DocumentTypeTransfer:
		return "Перевод"
	case deanery.DocumentTypeAcademicLeave:
		return "Академический отпуск"
	default:
		return string(docType)
	}
}

func (h *DeaneryHandler) getStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "approved":
		return "✅"
	case "rejected":
		return "❌"
	case "completed":
		return "✅"
	default:
		return "📄"
	}
}

func (h *DeaneryHandler) getStatusLabel(status string) string {
	switch status {
	case "pending":
		return "Ожидает обработки"
	case "approved":
		return "Одобрено"
	case "rejected":
		return "Отклонено"
	case "completed":
		return "Завершено"
	default:
		return status
	}
}

