package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Service определяет интерфейс для работы с AI
type Service interface {
	AskQuestion(ctx context.Context, question string, contextData ContextData) (string, error)
}

// ContextData содержит контекстную информацию о пользователе
type ContextData struct {
	UserInfo    UserInfo    `json:"user_info"`
	Schedule    []ScheduleItem `json:"schedule"`
	Courses     []Course    `json:"courses"`
}

// UserInfo содержит информацию о пользователе
type UserInfo struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

// ScheduleItem представляет элемент расписания
type ScheduleItem struct {
	Subject   string    `json:"subject"`
	Time      string    `json:"time"`
	Date      time.Time `json:"date"`
	Location  string    `json:"location"`
	Teacher   string    `json:"teacher"`
}

// Course представляет курс из Moodle
type Course struct {
	Fullname    string `json:"fullname"`
	Description string `json:"description"`
	StartDate   int64  `json:"startdate"`
	EndDate     int64  `json:"enddate"`
	Progress    *int   `json:"progress"`
	Completed   bool   `json:"completed"`
}

type YandexGPTService struct {
	APIKey   string
	APIURL   string
	FolderID string
	client   *http.Client
}

type YandexGPTRequest struct {
	ModelURI          string            `json:"modelUri"`
	CompletionOptions CompletionOptions `json:"completionOptions"`
	Messages          []Message         `json:"messages"`
}

type CompletionOptions struct {
	Stream           bool             `json:"stream"`
	Temperature      float64          `json:"temperature"`
	MaxTokens        string           `json:"maxTokens"`
	ReasoningOptions ReasoningOptions `json:"reasoningOptions"`
}

type ReasoningOptions struct {
	Mode string `json:"mode"`
}

type Message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type YandexGPTResponse struct {
	Result Result `json:"result"`
}

type Result struct {
	Alternatives []Alternative `json:"alternatives"`
	Usage        Usage         `json:"usage"`
	ModelVersion string        `json:"modelVersion"`
}

type Alternative struct {
	Message Message `json:"message"`
	Status  string  `json:"status"`
}

type Usage struct {
	InputTextTokens  string `json:"inputTextTokens"`
	CompletionTokens string `json:"completionTokens"`
	TotalTokens      string `json:"totalTokens"`
}

// NewYandexGPTService создает новый сервис YandexGPT
func NewYandexGPTService(apiKey, folderID string) Service {
	return &YandexGPTService{
		APIKey:   apiKey,
		APIURL:   "https://llm.api.cloud.yandex.net/foundationModels/v1/completion",
		FolderID: folderID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AskQuestion отправляет вопрос пользователя в YandexGPT с контекстом
func (s *YandexGPTService) AskQuestion(ctx context.Context, question string, contextData ContextData) (string, error) {
	// Формируем промпт с контекстом
	prompt := s.buildPrompt(question, contextData)

	// Отправляем запрос к YandexGPT
	response, err := s.callYandexGPT(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to call YandexGPT: %w", err)
	}

	// Очищаем ответ от markdown и лишних символов
	cleaned := s.cleanResponse(response)

	return cleaned, nil
}

// buildPrompt создает промпт для YandexGPT с контекстом пользователя
func (s *YandexGPTService) buildPrompt(question string, contextData ContextData) string {
	var contextParts []string

	// Информация о пользователе
	contextParts = append(contextParts, fmt.Sprintf("Информация о пользователе:\n- Имя: %s %s\n- Возраст: %d\n- Пол: %s\n- Email: %s\n- Роль: %s",
		contextData.UserInfo.FirstName,
		contextData.UserInfo.LastName,
		contextData.UserInfo.Age,
		contextData.UserInfo.Gender,
		contextData.UserInfo.Email,
		contextData.UserInfo.Role,
	))

	// Расписание
	if len(contextData.Schedule) > 0 {
		contextParts = append(contextParts, "\nРасписание пользователя:")
		for _, item := range contextData.Schedule {
			contextParts = append(contextParts, fmt.Sprintf("- %s в %s (%s, %s)", item.Subject, item.Time, item.Location, item.Teacher))
		}
	} else {
		contextParts = append(contextParts, "\nРасписание: пока нет данных")
	}

	// Курсы из Moodle
	if len(contextData.Courses) > 0 {
		contextParts = append(contextParts, "\nКурсы пользователя:")
		for _, course := range contextData.Courses {
			progress := "не указан"
			if course.Progress != nil {
				progress = fmt.Sprintf("%d%%", *course.Progress)
			}
			status := "в процессе"
			if course.Completed {
				status = "завершен"
			}
			contextParts = append(contextParts, fmt.Sprintf("- %s (прогресс: %s, статус: %s)", course.Fullname, progress, status))
		}
	} else {
		contextParts = append(contextParts, "\nКурсы: пока нет данных")
	}

	contextStr := strings.Join(contextParts, "\n")

	prompt := fmt.Sprintf(`Ты - умный помощник для студентов и сотрудников университета. Ты помогаешь отвечать на вопросы, связанные с учебным процессом, расписанием, курсами и другими аспектами университетской жизни.

Контекст о пользователе:
%s

Вопрос пользователя: %s

Ответь на вопрос пользователя, используя предоставленный контекст. Если в контексте нет нужной информации, ответь честно, что у тебя нет этой информации, но можешь дать общий совет. Отвечай на русском языке, дружелюбно и по делу.`, contextStr, question)

	return prompt
}

// cleanResponse очищает ответ от markdown и лишних символов
func (s *YandexGPTService) cleanResponse(content string) string {
	trimmed := strings.TrimSpace(content)

	// Удаляем markdown кодовые блоки
	trimmed = strings.ReplaceAll(trimmed, "```json", "")
	trimmed = strings.ReplaceAll(trimmed, "```JSON", "")
	trimmed = strings.ReplaceAll(trimmed, "```", "")
	trimmed = strings.TrimSpace(trimmed)

	return trimmed
}

// callYandexGPT отправляет запрос к YandexGPT API
func (s *YandexGPTService) callYandexGPT(ctx context.Context, prompt string) (string, error) {
	request := YandexGPTRequest{
		ModelURI: fmt.Sprintf("gpt://%s/yandexgpt-lite/latest", s.FolderID),
		CompletionOptions: CompletionOptions{
			Stream:      false,
			Temperature: 0.6,
			MaxTokens:   "2000",
			ReasoningOptions: ReasoningOptions{
				Mode: "DISABLED",
			},
		},
		Messages: []Message{
			{
				Role: "system",
				Text: "Ты - умный помощник для студентов и сотрудников университета. Ты помогаешь отвечать на вопросы, связанные с учебным процессом, расписанием, курсами и другими аспектами университетской жизни. Отвечай на русском языке, дружелюбно и по делу.",
			},
			{
				Role: "user",
				Text: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("x-folder-id", s.FolderID)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("YandexGPT API returned status: %d", resp.StatusCode)
	}

	var response YandexGPTResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if len(response.Result.Alternatives) == 0 {
		return "", fmt.Errorf("no response from YandexGPT")
	}

	return response.Result.Alternatives[0].Message.Text, nil
}

