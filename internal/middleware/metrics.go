package middleware

import (
	"net/http"
	"gosprints/internal/metrics"
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Игнорируем эндпоинты метрик
		if r.URL.Path == "/metrics" || r.URL.Path == "/metrics/prometheus" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Считаем запрос
		metrics.Get().IncAPIRequests()
		
		// Обертка для отслеживания статусов
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		if rw.statusCode >= 400 {
			metrics.Get().IncAPIErrors()
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}