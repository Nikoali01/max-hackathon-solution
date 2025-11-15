package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"github.com/rs/zerolog"

	"first-max-bot/internal/bot"
	"first-max-bot/internal/services/deanery"
	"first-max-bot/internal/services/user"
	"first-max-bot/internal/state"
)

// DocumentsHandler обрабатывает команду /documents для администраторов
type DocumentsHandler struct {
	deaneryService deanery.Service
	userService    user.Service
	logger         zerolog.Logger
}

func NewDocumentsHandler(deaneryService deanery.Service, userService user.Service, logger zerolog.Logger) *DocumentsHandler {
	return &DocumentsHandler{
		deaneryService: deaneryService,
		userService:    userService,
		logger:        logger,
	}
}

func (h *DocumentsHandler) Handle(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	// Проверяем, что пользователь - руководитель
	userID := req.UserID()
	u, err := h.userService.GetUserByID(ctx, userID)
	if err != nil || u == nil || u.Role != user.RoleManager {
		return responder.SendText(ctx, req.Recipient(), "❌ Эта команда доступна только руководителям.")
	}

	// Если это текстовый ввод для ответа на заявление (включая сообщения с файлами)
	userState := req.UserState
	if userState != nil && userState.UserRegistrationStep == "doc_response" {
		// Проверяем, есть ли файл или текст
		hasText := strings.TrimSpace(req.Args) != ""
		hasFile := false
		if req.Update != nil {
			hasFile = (req.Update.Message.Body.RawAttachments != nil && len(req.Update.Message.Body.RawAttachments) > 0) ||
				(req.Update.Message.Body.Attachments != nil && len(req.Update.Message.Body.Attachments) > 0)
		}
		
		// Если есть текст или файл, обрабатываем как ответ
		if hasText || hasFile {
			return h.HandleTextInput(ctx, req, responder)
		}
		// Если нет ни текста, ни файла, но состояние doc_response, все равно обрабатываем
		// (может быть пустое сообщение, которое нужно обработать)
		return h.HandleTextInput(ctx, req, responder)
	}

	// Если это callback для просмотра или ответа на заявление
	if strings.HasPrefix(req.Args, "doc_admin:") {
		return h.handleDocumentCallback(ctx, req, responder)
	}

	// Получаем все заявления
	documents, err := h.deaneryService.GetAllDocuments(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get documents")
		return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при получении заявлений")
	}

	// Фильтруем необработанные заявления
	var pendingDocs []deanery.Document
	for _, doc := range documents {
		if doc.Status == "pending" {
			pendingDocs = append(pendingDocs, doc)
		}
	}

	var message strings.Builder
	message.WriteString("📋 Заявления деканата\n\n")

	if len(pendingDocs) == 0 {
		message.WriteString("✅ Нет необработанных заявлений.")
		return responder.SendText(ctx, req.Recipient(), message.String())
	}

	message.WriteString(fmt.Sprintf("Необработанных заявлений: %d\n\n", len(pendingDocs)))

	keyboard := responder.NewKeyboardBuilder()

	// Показываем первые 10 заявлений
	for i, doc := range pendingDocs {
		if i >= 10 {
			break
		}

		row := keyboard.AddRow()
		docTypeLabel := h.getDocumentTypeLabel(doc.Type)
		subject := fmt.Sprintf("%s #%s", docTypeLabel, doc.ID)
		if len(subject) > 30 {
			subject = subject[:30] + "..."
		}
		row.AddCallback(fmt.Sprintf("📄 %s", subject), schemes.DEFAULT, fmt.Sprintf("doc_admin:view:%s", doc.ID))
	}

	return responder.SendTextWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
}

func (h *DocumentsHandler) handleDocumentCallback(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	payload := req.Args

	callbackID := ""
	if req.Metadata != nil {
		if cid, ok := req.Metadata["callback_id"].(string); ok && cid != "" {
			callbackID = cid
			responder.AnswerCallback(ctx, callbackID, &schemes.CallbackAnswer{})
		}
	}

	if strings.HasPrefix(payload, "doc_admin:view:") {
		docID := strings.TrimPrefix(payload, "doc_admin:view:")

		doc, err := h.deaneryService.GetDocument(ctx, docID)
		if err != nil || doc == nil {
			return responder.SendText(ctx, req.Recipient(), "❌ Заявление не найдено")
		}

		var message strings.Builder
		message.WriteString(fmt.Sprintf("📄 Заявление #%s\n\n", doc.ID))
		message.WriteString(fmt.Sprintf("Тип: %s\n", h.getDocumentTypeLabel(doc.Type)))
		message.WriteString(fmt.Sprintf("От пользователя: %s\n", doc.UserID))
		message.WriteString(fmt.Sprintf("Статус: %s\n", doc.Status))
		message.WriteString(fmt.Sprintf("Создано: %s\n\n", doc.CreatedAt.Format("02.01.2006 15:04")))
		message.WriteString(fmt.Sprintf("Описание:\n%s\n\n", doc.Description))

		if doc.Response != "" {
			message.WriteString(fmt.Sprintf("📤 Ответ:\n%s\n\n", doc.Response))
		} else {
			message.WriteString("⏳ Ожидает обработки\n\n")
		}

		keyboard := responder.NewKeyboardBuilder()
		if doc.Status == "pending" {
			row := keyboard.AddRow()
			row.AddCallback("✍️ Ответить", schemes.POSITIVE, fmt.Sprintf("doc_admin:reply:%s", doc.ID))
		}

		return responder.SendTextWithKeyboard(ctx, req.Recipient(), message.String(), keyboard)
	}

	if strings.HasPrefix(payload, "doc_admin:reply:") {
		docID := strings.TrimPrefix(payload, "doc_admin:reply:")

		// Сохраняем docID в состоянии для ответа
		if req.UserState == nil {
			req.UserState = &state.UserState{
				UserRegistrationData: make(map[string]string),
			}
		}
		if req.UserState.UserRegistrationData == nil {
			req.UserState.UserRegistrationData = make(map[string]string)
		}
		req.UserState.UserRegistrationData["replying_to_doc"] = docID
		req.UserState.UserRegistrationStep = "doc_response"

		message := "✍️ Напиши ответ на заявление:\n\n"
		message += "Отправь либо только текст, либо только файл (нельзя отправлять и то, и другое одновременно)."

		return responder.SendText(ctx, req.Recipient(), message)
	}

	return nil
}

func (h *DocumentsHandler) HandleTextInput(ctx context.Context, req *bot.Request, responder bot.Responder) error {
	userState := req.UserState

	if userState != nil && userState.UserRegistrationStep == "doc_response" {
		docID := userState.UserRegistrationData["replying_to_doc"]
		if docID == "" {
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка: не найден ID заявления")
		}

		responseText := strings.TrimSpace(req.Args)

		// Получаем ID текущего администратора
		adminUserID := req.UserID()

		// Обработка файлов из сообщения
		responseFile := ""
		if req.Update != nil {
			// Сначала проверяем RawAttachments (сырые JSON данные)
			if req.Update.Message.Body.RawAttachments != nil {
				h.logger.Debug().Int("count", len(req.Update.Message.Body.RawAttachments)).Msg("checking RawAttachments")
				for _, rawAtt := range req.Update.Message.Body.RawAttachments {
					var att map[string]interface{}
					if err := json.Unmarshal(rawAtt, &att); err != nil {
						h.logger.Debug().Err(err).Msg("failed to unmarshal raw attachment")
						continue
					}
					
					// Проверяем тип attachment
					attType, ok := att["type"].(string)
					if !ok {
						continue
					}
					h.logger.Debug().Str("type", attType).Msg("found attachment type")
					
					if attType != "file" {
						continue
					}
					
					// Для FileAttachment payload содержит token или url
					payload, ok := att["payload"].(map[string]interface{})
					if !ok {
						h.logger.Debug().Msg("payload not found or not a map")
						continue
					}
					
					// Используем token если есть, иначе url
					if token, ok := payload["token"].(string); ok && token != "" {
						responseFile = token
						h.logger.Info().Str("token", token).Msg("found file token")
						break
					} else if url, ok := payload["url"].(string); ok && url != "" {
						responseFile = url
						h.logger.Info().Str("url", url).Msg("found file url")
						break
					}
				}
			}
			
			// Также проверяем Attachments (обработанные attachments)
			if responseFile == "" && req.Update.Message.Body.Attachments != nil {
				h.logger.Debug().Int("count", len(req.Update.Message.Body.Attachments)).Msg("checking Attachments")
				for i, att := range req.Update.Message.Body.Attachments {
					if fileAtt, ok := att.(map[string]interface{}); ok {
						attType, ok := fileAtt["type"].(string)
						if !ok || attType != "file" {
							continue
						}
						
						h.logger.Debug().Int("index", i).Str("type", attType).Msg("found file attachment")
						
						payload, ok := fileAtt["payload"].(map[string]interface{})
						if !ok {
							continue
						}
						
						if token, ok := payload["token"].(string); ok && token != "" {
							responseFile = token
							h.logger.Info().Str("token", token).Msg("found file token from Attachments")
							break
						} else if url, ok := payload["url"].(string); ok && url != "" {
							responseFile = url
							h.logger.Info().Str("url", url).Msg("found file url from Attachments")
							break
						}
					} else {
						// Попробуем привести к FileAttachment напрямую
						if fileAttStruct, ok := att.(*schemes.FileAttachment); ok {
							if fileAttStruct.Payload.Token != "" {
								responseFile = fileAttStruct.Payload.Token
								h.logger.Info().Str("token", responseFile).Msg("found file token from FileAttachment struct")
								break
							} else if fileAttStruct.Payload.Url != "" {
								responseFile = fileAttStruct.Payload.Url
								h.logger.Info().Str("url", responseFile).Msg("found file url from FileAttachment struct")
								break
							}
						}
					}
				}
			}
		}
		
		h.logger.Debug().Str("responseText", responseText).Str("responseFile", responseFile).Msg("extracted response data")

		// Проверяем, что есть либо текст, либо файл (но не оба одновременно)
		if responseText == "" && responseFile == "" {
			return responder.SendText(ctx, req.Recipient(), "❌ Ответ не может быть пустым. Отправь либо текст, либо файл.")
		}
		
		// Проверяем, что не отправлены и текст, и файл одновременно
		if responseText != "" && responseFile != "" {
			return responder.SendText(ctx, req.Recipient(), "❌ Можно отправить либо только текст, либо только файл. Нельзя отправлять и то, и другое одновременно.")
		}

		err := h.deaneryService.AddDocumentResponse(ctx, docID, responseText, responseFile, adminUserID)
		if err != nil {
			h.logger.Error().Err(err).Msg("failed to add document response")
			return responder.SendText(ctx, req.Recipient(), "❌ Ошибка при сохранении ответа")
		}

		// Получаем документ для отправки уведомления пользователю
		doc, err := h.deaneryService.GetDocument(ctx, docID)
		if err == nil && doc != nil {
			// Отправляем уведомление пользователю
			userIDInt, err := strconv.ParseInt(doc.UserID, 10, 64)
			if err == nil {
				userRecipient := schemes.Recipient{
					UserId:   userIDInt,
					ChatType: schemes.DIALOG,
				}

				notification := fmt.Sprintf("✅ Ответ на твоё заявление #%s\n\n", docID)
				notification += fmt.Sprintf("Тип: %s\n\n", h.getDocumentTypeLabel(doc.Type))
				if responseText != "" {
					notification += fmt.Sprintf("Ответ:\n%s\n\n", responseText)
				}
				if responseFile != "" {
					notification += "📎 К заявлению приложен файл.\n\n"
				}
				notification += "Используй /deanery чтобы посмотреть все свои заявления."

				// Отправляем уведомление пользователю с файлом (если есть)
				if responseFile != "" {
					// Отправляем текст с файлом
					if err := responder.SendTextWithFile(ctx, userRecipient, notification, responseFile); err != nil {
						h.logger.Warn().Err(err).Str("user_id", doc.UserID).Msg("failed to send notification with file to user")
					} else {
						h.logger.Info().Str("doc_id", docID).Str("user_id", doc.UserID).Str("file_token", responseFile).Msg("user notified about document response with file")
					}
				} else {
					// Отправляем только текст
					if err := responder.SendText(ctx, userRecipient, notification); err != nil {
						h.logger.Warn().Err(err).Str("user_id", doc.UserID).Msg("failed to send notification to user")
					} else {
						h.logger.Info().Str("doc_id", docID).Str("user_id", doc.UserID).Msg("user notified about document response")
					}
				}
			}
		}

		// Очищаем состояние
		userState.UserRegistrationStep = ""
		delete(userState.UserRegistrationData, "replying_to_doc")

		message := fmt.Sprintf("✅ Ответ на заявление #%s сохранён!\n\n", docID)
		if responseText != "" {
			message += fmt.Sprintf("Ответ:\n%s\n\n", responseText)
		}
		if responseFile != "" {
			message += "📎 Файл приложен.\n\n"
		}
		message += "Пользователь получит уведомление."

		return responder.SendText(ctx, req.Recipient(), message)
	}

	return nil
}

func (h *DocumentsHandler) getDocumentTypeLabel(docType deanery.DocumentType) string {
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

