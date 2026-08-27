package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	
	
	"github.com/google/uuid"
	
	"github.com/islamova-h/crawler-service/internal/crawler"
	"github.com/islamova-h/crawler-service/internal/logger"
	"github.com/islamova-h/crawler-service/internal/models"
)

type CrawlerHandler struct {
	pool *crawler.WorkerPool
}

func NewCrawlerHandler(pool *crawler.WorkerPool) *CrawlerHandler {
	return &CrawlerHandler{
		pool: pool,
	}
}

// Crawl обрабатывает POST /crawl
func (h *CrawlerHandler) Crawl(w http.ResponseWriter, r *http.Request) {
	// Генерируем trace_id для всего запроса
	traceID := uuid.New().String()
	ctx := logger.NewContextWithTraceID(r.Context(), traceID)
	
	// Логируем входящий запрос
	logger.Info(ctx, "Received crawl request", map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})
	
	// Проверяем метод
	if r.Method != http.MethodPost {
		logger.Warn(ctx, "Invalid method", map[string]interface{}{
			"method": r.Method,
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Читаем тело
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(ctx, "Failed to read body", map[string]interface{}{
			"error": err.Error(),
		})
		sendError(w, "Failed to read request body", traceID, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Парсим JSON
	var req models.CrawlRequest
	if err := json.Unmarshal(body, &req); err != nil {
		logger.Error(ctx, "Failed to parse JSON", map[string]interface{}{
			"error": err.Error(),
		})
		sendError(w, "Invalid JSON format", traceID, http.StatusBadRequest)
		return
	}
	
	// Валидируем
	if len(req.URLs) == 0 {
		logger.Warn(ctx, "Empty URLs list", nil)
		sendError(w, "URLs list cannot be empty", traceID, http.StatusBadRequest)
		return
	}
	
	// Создаем контекст с таймаутом для всего запроса (опционально)
	// Если клиент отменит запрос, контекст отменится автоматически
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Важно! Освобождаем ресурсы
	
	// Запускаем краулинг
	results := h.pool.Crawl(ctx, req.URLs)
	
	// Формируем ответ
	resp := models.CrawlResponse{
		Results: results,
	}
	
	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error(ctx, "Failed to encode response", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	
	logger.Info(ctx, "Crawl request completed", map[string]interface{}{
		"results_count": len(results),
	})
}

func sendError(w http.ResponseWriter, message, traceID string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   message,
		TraceID: traceID,
	})
}