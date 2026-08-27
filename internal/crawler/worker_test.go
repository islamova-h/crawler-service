package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	//"github.com/islamova-h/crawler-service/internal/models"
)

func TestWorkerPool_Crawl_Success(t *testing.T) {
	// Создаем тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test body"))
	}))
	defer ts.Close()

	client := NewClient(3 * time.Second)
	pool := NewWorkerPool(2, client)

	urls := []string{ts.URL, ts.URL, ts.URL}
	results := pool.Crawl(context.Background(), urls)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Status != http.StatusOK {
			t.Errorf("Expected status 200, got %d", result.Status)
		}
		if result.Length != 9 { // "test body" = 9 байт
			t.Errorf("Expected length 9, got %d", result.Length)
		}
	}
}

func TestWorkerPool_Crawl_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(5 * time.Second)
	pool := NewWorkerPool(2, client)

	ctx, cancel := context.WithCancel(context.Background())
	
	// Отменяем контекст через 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	urls := []string{ts.URL, ts.URL, ts.URL}
	results := pool.Crawl(ctx, urls)

	// Может вернуть 0 или часть результатов - это нормально
	// Главное, чтобы не упало
	t.Logf("Got %d results after cancellation", len(results))
}

func TestWorkerPool_Crawl_EmptyURLs(t *testing.T) {
	client := NewClient(3 * time.Second)
	pool := NewWorkerPool(2, client)

	results := pool.Crawl(context.Background(), []string{})

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}