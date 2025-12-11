package handlers

import (
	"encoding/json"
	"net/http"

	"gosprints/internal/services"
)

type CacheHandler struct {
	taskRepo services.TaskCacheRepository
}

func NewCacheHandler(taskRepo services.TaskCacheRepository) *CacheHandler {
	return &CacheHandler{taskRepo: taskRepo}
}

func (h *CacheHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := h.taskRepo.GetCacheStats()
	
	hitRate := 0.0
	if stats.Hits+stats.Misses > 0 {
		hitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
	}
	
	response := map[string]interface{}{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"sets":        stats.Sets,
		"deletes":     stats.Deletes,
		"expirations": stats.Expirations,
		"hit_rate":    hitRate,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CacheHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	err := h.taskRepo.ClearCache(r.Context())
	if err != nil {
		http.Error(w, "Failed to clear cache", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Cache cleared successfully",
	})
}

func (h *CacheHandler) WarmUpCache(w http.ResponseWriter, r *http.Request) {
	err := h.taskRepo.WarmUpCache(r.Context())
	if err != nil {
		http.Error(w, "Failed to warm up cache", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Cache warmed up successfully",
	})
}