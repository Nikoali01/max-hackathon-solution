package moodle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	MoodleBaseURL = "http://95.81.124.161:80"
)

// SiteInfo представляет информацию о сайте Moodle и пользователе
type SiteInfo struct {
	Sitename        string   `json:"sitename"`
	Username        string   `json:"username"`
	Firstname       string   `json:"firstname"`
	Lastname        string   `json:"lastname"`
	Fullname        string   `json:"fullname"`
	Lang            string   `json:"lang"`
	UserID          int      `json:"userid"`
	SiteURL         string   `json:"siteurl"`
	UserPictureURL  string   `json:"userpictureurl"`
	Functions       []Function `json:"functions"`
	Release         string   `json:"release"`
	Version         string   `json:"version"`
	UserCanManageOwnFiles bool `json:"usercanmanageownfiles"`
	UserQuota       int      `json:"userquota"`
	UserMaxUploadFileSize int `json:"usermaxuploadfilesize"`
}

// Course представляет курс Moodle
type Course struct {
	ID              int    `json:"id"`
	Shortname       string `json:"shortname"`
	Fullname        string `json:"fullname"`
	Displayname     string `json:"displayname"`
	Summary         string `json:"summary"` // HTML описание
	StartDate       int64  `json:"startdate"` // Unix timestamp
	EndDate         int64  `json:"enddate"`   // Unix timestamp
	Progress        *int   `json:"progress"`  // Может быть null
	Completed       bool   `json:"completed"`
	LastAccess      int64  `json:"lastaccess"` // Unix timestamp
	EnrolledUserCount int  `json:"enrolledusercount"`
	CourseImage     string `json:"courseimage"`
}

// Function представляет доступную функцию Moodle API
type Function struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Service interface {
	GetSiteInfo(ctx context.Context, token string) (*SiteInfo, error)
	GetUserCourses(ctx context.Context, token string, userID int) ([]Course, error)
}

type httpService struct {
	baseURL string
	client  *http.Client
}

func NewService() Service {
	return &httpService{
		baseURL: MoodleBaseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *httpService) GetSiteInfo(ctx context.Context, token string) (*SiteInfo, error) {
	// Формируем URL для запроса
	apiURL := fmt.Sprintf("%s/webservice/rest/server.php", s.baseURL)
	
	params := url.Values{}
	params.Set("wstoken", token)
	params.Set("wsfunction", "core_webservice_get_site_info")
	params.Set("moodlewsrestformat", "json")
	
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
	
	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Выполняем запрос
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("moodle API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Парсим ответ
	var siteInfo SiteInfo
	if err := json.NewDecoder(resp.Body).Decode(&siteInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &siteInfo, nil
}

func (s *httpService) GetUserCourses(ctx context.Context, token string, userID int) ([]Course, error) {
	// Формируем URL для запроса
	apiURL := fmt.Sprintf("%s/webservice/rest/server.php", s.baseURL)
	
	params := url.Values{}
	params.Set("wstoken", token)
	params.Set("wsfunction", "core_enrol_get_users_courses")
	params.Set("moodlewsrestformat", "json")
	params.Set("userid", fmt.Sprintf("%d", userID))
	
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
	
	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Выполняем запрос
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("moodle API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Парсим ответ
	var courses []Course
	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return courses, nil
}

