package crawler

import (
	"context"
	"sync"
	"time"
	
	"github.com/islamova-h/crawler-service/internal/logger"
	"github.com/islamova-h/crawler-service/internal/models"
)

// WorkerPool управляет пулом воркеров
type WorkerPool struct {
	workers int
	client  *Client
}

// NewWorkerPool создает новый пул
func NewWorkerPool(workers int, client *Client) *WorkerPool {
	if workers <= 0 {
		workers = 5
	}
	return &WorkerPool{
		workers: workers,
		client:  client,
	}
}

// Crawl обрабатывает список URL-ов
func (wp *WorkerPool) Crawl(ctx context.Context, urls []string) []models.CrawlResult {
	if len(urls) == 0 {
		return []models.CrawlResult{}
	}
	
	startTime := time.Now()
	
	// Логируем старт
	logger.Info(ctx, "Starting crawl", map[string]interface{}{
		"url_count": len(urls),
		"workers":   wp.workers,
	})
	
	// Создаем каналы
	jobs := make(chan string, len(urls))
	results := make(chan models.CrawlResult, len(urls))
	
	// WaitGroup для ожидания завершения всех воркеров
	var wg sync.WaitGroup
	
	// Запускаем воркеры
	for i := 0; i < wp.workers; i++ {
		wg.Add(1)
		go wp.worker(ctx, i, jobs, results, &wg)
	}
	
	// Отправляем задачи
	go func() {
		for _, url := range urls {
			select {
			case <-ctx.Done():
				// Контекст отменен - выходим
				logger.Warn(ctx, "Context cancelled while sending jobs", map[string]interface{}{
					"remaining": len(urls),
				})
				close(jobs)
				return
			case jobs <- url:
			}
		}
		close(jobs)
	}()
	
	// Ждем завершения воркеров в отдельной горутине
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Собираем результаты
	var resultsSlice []models.CrawlResult
	for result := range results {
		resultsSlice = append(resultsSlice, result)
	}
	
	duration := time.Since(startTime).Milliseconds()
	logger.Info(ctx, "Crawl completed", map[string]interface{}{
		"processed":   len(resultsSlice),
		"duration_ms": duration,
	})
	
	return resultsSlice
}

// worker - горутина, обрабатывающая задачи
func (wp *WorkerPool) worker(ctx context.Context, id int, jobs <-chan string, results chan<- models.CrawlResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for url := range jobs {
		// Проверяем, не отменен ли контекст
		select {
		case <-ctx.Done():
			logger.Warn(ctx, "Worker stopped due to context cancellation", map[string]interface{}{
				"worker_id": id,
				"url":       url,
			})
			return
		default:
			// Продолжаем работу
		}
		
		result := models.CrawlResult{URL: url}
		startTime := time.Now()
		
		// Делаем запрос
		status, length, err := wp.client.Fetch(ctx, url)
		
		if err != nil {
			// Определяем статус ошибки
			if ctx.Err() != nil {
				// Контекст отменен - выходим
				logger.Warn(ctx, "Context cancelled during fetch", map[string]interface{}{
					"worker_id": id,
					"url":       url,
					"error":     err.Error(),
				})
				return
			}
			
			// Таймаут или другая ошибка
			result.Status = 503 // Service Unavailable
			if err.Error() == "context deadline exceeded" {
				result.Status = 503
			} else {
				result.Status = 500
			}
			result.Length = 0
			
			logger.Warn(ctx, "Failed to fetch URL", map[string]interface{}{
				"worker_id":   id,
				"url":         url,
				"status":      result.Status,
				"error":       err.Error(),
				"duration_ms": time.Since(startTime).Milliseconds(),
			})
		} else {
			result.Status = status
			result.Length = length
			
			logger.Info(ctx, "Fetched URL", map[string]interface{}{
				"worker_id":   id,
				"url":         url,
				"status":      status,
				"length":      length,
				"duration_ms": time.Since(startTime).Milliseconds(),
			})
		}
		
		// Отправляем результат (не блокируем, если контекст отменен)
		select {
		case <-ctx.Done():
			logger.Warn(ctx, "Context cancelled, dropping result", map[string]interface{}{
				"worker_id": id,
				"url":       url,
			})
			return
		case results <- result:
		}
	}
}