package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	
	"gosprints/internal/metrics"
)

type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	hijacked   bool
}

func (w *wrappedResponseWriter) WriteHeader(statusCode int) {
	if !w.hijacked {
		w.statusCode = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *wrappedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		w.hijacked = true
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

func (w *wrappedResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Для всех запросов
		metrics.Get().IncAPIRequests()
		
		// Создаем обертку
		wrapped := &wrappedResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Выполняем handler
		next.ServeHTTP(wrapped, r)
		
		// Если это был не WebSocket (не был hijacked), считаем ошибки
		if !wrapped.hijacked && wrapped.statusCode >= 400 {
			metrics.Get().IncAPIErrors()
		}
	})
}