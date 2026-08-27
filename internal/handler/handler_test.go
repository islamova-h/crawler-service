package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/islamova-h/crawler-service/internal/crawler"
	"github.com/islamova-h/crawler-service/internal/models"
)

func TestCrawlHandler_Success(t *testing.T) {
	// Создаем тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer ts.Close()

	client := crawler.NewClient(3 * time.Second)
	pool := crawler.NewWorkerPool(2, client)
	handler := NewCrawlerHandler(pool)

	// Создаем запрос
	reqBody := models.CrawlRequest{
		URLs: []string{ts.URL, ts.URL},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/crawl", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем хендлер
	handler.Crawl(w, req)

	// Проверяем ответ
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp models.CrawlResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(resp.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(resp.Results))
	}
}

func TestCrawlHandler_EmptyURLs(t *testing.T) {
	client := crawler.NewClient(3 * time.Second)
	pool := crawler.NewWorkerPool(2, client)
	handler := NewCrawlerHandler(pool)

	reqBody := models.CrawlRequest{
		URLs: []string{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/crawl", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Crawl(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCrawlHandler_InvalidJSON(t *testing.T) {
	client := crawler.NewClient(3 * time.Second)
	pool := crawler.NewWorkerPool(2, client)
	handler := NewCrawlerHandler(pool)

	req := httptest.NewRequest(http.MethodPost, "/crawl", bytes.NewReader([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Crawl(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCrawlHandler_WrongMethod(t *testing.T) {
	client := crawler.NewClient(3 * time.Second)
	pool := crawler.NewWorkerPool(2, client)
	handler := NewCrawlerHandler(pool)

	req := httptest.NewRequest(http.MethodGet, "/crawl", nil)
	w := httptest.NewRecorder()

	handler.Crawl(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}