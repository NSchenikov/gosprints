package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gosprints/internal/cache"
	"gosprints/internal/metrics"
	"gosprints/internal/repositories"
	"gosprints/internal/ws"
	"gosprints/internal/services"
)

type MetricsHandler struct {
	hub     *ws.NotificationHub
	taskRepo services.TaskRepository
	cache   cache.Cache
}

func NewMetricsHandler(
	hub *ws.NotificationHub, 
	taskRepo services.TaskRepository,
	cache cache.Cache,
) *MetricsHandler {
	return &MetricsHandler{
		hub:     hub,
		taskRepo: taskRepo,
		cache:   cache,
	}
}

func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	m := metrics.Get()
	
	var cacheStats interface{} = nil
	if cacheRepo, ok := h.taskRepo.(interface{ GetCacheStats() cache.CacheStats }); ok {
		stats := cacheRepo.GetCacheStats()
		cacheStats = map[string]interface{}{
			"hits":        stats.Hits,
			"misses":      stats.Misses,
			"sets":        stats.Sets,
			"deletes":     stats.Deletes,
			"expirations": stats.Expirations,
			"hit_rate":    stats.HitRate(),
		}
	}
	
	response := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    m.Uptime().String(),
		
		"tasks": map[string]interface{}{
			"created":   m.TasksTotal(),
			"completed": m.TasksCompleted,
			"failed":    m.TasksFailed,
		},
		
		"processing": map[string]interface{}{
			"avg_time_ms": m.AvgProcessingTime().Milliseconds(),
			"total_tasks": m.processingCount,
		},
		
		"websocket": map[string]interface{}{
			"connections_active": m.GetWSConnectionsActive(),
			"connections_total":  m.wsConnectionsTotal,
		},
		
		"api": map[string]interface{}{
			"requests_total": m.apiRequestsTotal,
			"errors_total":   m.apiErrorsTotal,
			"error_rate":     m.APIErrorRate(),
		},
	}
	
	if cacheStats != nil {
		response["cache"] = cacheStats
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *MetricsHandler) GetPrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	m := metrics.Get()
	
	var cacheHits, cacheMisses, cacheSets int64
	if cacheRepo, ok := h.taskRepo.(interface{ GetCacheStats() cache.CacheStats }); ok {
		stats := cacheRepo.GetCacheStats()
		cacheHits = stats.Hits
		cacheMisses = stats.Misses
		cacheSets = stats.Sets
	}
	
	prometheus := fmt.Sprintf(`# HELP sprints_tasks_created_total Total tasks created
# TYPE sprints_tasks_created_total counter
sprints_tasks_created_total %d

# HELP sprints_tasks_completed_total Total tasks completed
# TYPE sprints_tasks_completed_total counter
sprints_tasks_completed_total %d

# HELP sprints_processing_time_avg_ms Average task processing time in milliseconds
# TYPE sprints_processing_time_avg_ms gauge
sprints_processing_time_avg_ms %.2f

# HELP sprints_ws_connections_active Active WebSocket connections
# TYPE sprints_ws_connections_active gauge
sprints_ws_connections_active %d

# HELP sprints_ws_connections_total Total WebSocket connections
# TYPE sprints_ws_connections_total counter
sprints_ws_connections_total %d

# HELP sprints_api_requests_total Total API requests
# TYPE sprints_api_requests_total counter
sprints_api_requests_total %d

# HELP sprints_api_errors_total Total API errors
# TYPE sprints_api_errors_total counter
sprints_api_errors_total %d

# HELP sprints_cache_hits_total Cache hits
# TYPE sprints_cache_hits_total counter
sprints_cache_hits_total %d

# HELP sprints_cache_misses_total Cache misses
# TYPE sprints_cache_misses_total counter
sprints_cache_misses_total %d

# HELP sprints_cache_sets_total Cache sets
# TYPE sprints_cache_sets_total counter
sprints_cache_sets_total %d

# HELP sprints_uptime_seconds Application uptime in seconds
# TYPE sprints_uptime_seconds gauge
sprints_uptime_seconds %.0f
`,
		m.TasksTotal(),
		m.TasksCompleted,
		m.AvgProcessingTime().Seconds()*1000,
		m.GetWSConnectionsActive(),
		m.wsConnectionsTotal,
		m.apiRequestsTotal,
		m.apiErrorsTotal,
		cacheHits,
		cacheMisses,
		cacheSets,
		m.Uptime().Seconds(),
	)
	
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte(prometheus))
}