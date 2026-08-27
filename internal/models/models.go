package models

// CrawlRequest запрос от клиента
type CrawlRequest struct {
	URLs []string `json:"urls"`
}

// CrawlResult результат для одного URL
type CrawlResult struct {
	URL     string `json:"url"`
	Status  int    `json:"status"`
	Length  int    `json:"length"`
}

// CrawlResponse ответ сервера
type CrawlResponse struct {
	Results []CrawlResult `json:"results"`
}

// ErrorResponse ошибка
type ErrorResponse struct {
	Error   string `json:"error"`
	TraceID string `json:"trace_id,omitempty"`
}