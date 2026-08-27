package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout = 3 * time.Second
)

// HTTPClient интерфейс для тестирования
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client обертка над http.Client
type Client struct {
	httpClient HTTPClient
	timeout    time.Duration
}

// NewClient создает новый клиент
func NewClient(timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = defaultTimeout
	}
	
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
			// Важно: настраиваем транспорт для переиспользования соединений
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		timeout: timeout,
	}
}

// Fetch делает запрос к URL и возвращает статус и длину тела
func (c *Client) Fetch(ctx context.Context, url string) (status int, length int, err error) {
	// Создаем запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("create request: %w", err)
	}
	
	// Добавляем User-Agent (хороший тон)
	req.Header.Set("User-Agent", "CrawlerBot/1.0")
	
	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	
	// Читаем тело
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, 0, fmt.Errorf("read body: %w", err)
	}
	
	return resp.StatusCode, len(body), nil
}