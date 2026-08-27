package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Fetch_Success(t *testing.T) {
	// Создаем тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	defer ts.Close()

	client := NewClient(3 * time.Second)
	status, length, err := client.Fetch(context.Background(), ts.URL)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}
	if length != 13 {
		t.Errorf("Expected length 13, got %d", length)
	}
}

func TestClient_Fetch_Timeout(t *testing.T) {
	// Сервер, который долго отвечает
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(1 * time.Second) // Таймаут 1 секунда
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	status, _, err := client.Fetch(ctx, ts.URL)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	if status != 0 {
		t.Errorf("Expected status 0 on error, got %d", status)
	}
}

func TestClient_Fetch_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(5 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	
	// Отменяем контекст сразу
	cancel()

	_, _, err := client.Fetch(ctx, ts.URL)

	if err == nil {
		t.Error("Expected context cancelled error, got nil")
	}
}