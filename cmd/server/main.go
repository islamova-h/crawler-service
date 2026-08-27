package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/islamova-h/crawler-service/internal/crawler"
	"github.com/islamova-h/crawler-service/internal/handler"
	"github.com/islamova-h/crawler-service/internal/logger"
)

func main() {
	// Инициализируем зависимости
	client := crawler.NewClient(3 * time.Second)
	pool := crawler.NewWorkerPool(5, client)
	h := handler.NewCrawlerHandler(pool)
	
	// Настраиваем роутер
	mux := http.NewServeMux()
	mux.HandleFunc("POST /crawl", h.Crawl)
	
	// Добавляем middleware для логирования всех запросов
	handlerWithLogging := loggingMiddleware(mux)
	
	// Создаем сервер
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handlerWithLogging,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Server starting on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// Даем время на завершение текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited properly")
}

// loggingMiddleware логирует все HTTP запросы
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Генерируем trace_id если его нет
		traceID := logger.GetTraceID(r.Context())
		if traceID == "" {
			traceID = "no-trace"
		}
		
		log.Printf("[%s] %s %s", traceID, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}