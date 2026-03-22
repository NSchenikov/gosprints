package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"log"
	// "time"

	"task-service/internal/grpc/task/client"
)

type TaskHandler struct {
	taskClient *client.TaskClient
}

func NewTaskHandler(taskClient *client.TaskClient) *TaskHandler {
	return &TaskHandler{
		taskClient: taskClient,
	}
}

// GetTasks - получение всех задач пользователя
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	// userID должен приходить от api-gateway (например, в заголовке)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// Или из query параметра для обратной совместимости
		userID = r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}
	}
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	
	status := r.URL.Query().Get("status")
	
	tasks, total, err := h.taskClient.ListTasks(r.Context(), userID, status, int32(page), int32(pageSize))
	if err != nil {
		log.Printf("Error getting tasks: %v\n", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks":     tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateTask - создание задачи
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}
	}

	var input struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	task, err := h.taskClient.CreateTask(r.Context(), input.Text, userID)
	if err != nil {
		log.Printf("CreateTask error: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTaskByID - получение задачи по ID
func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}
	}
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.taskClient.GetTask(r.Context(), int32(id))
	if err != nil {
		log.Printf("Error getting task: %v\n", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	
	// Проверяем принадлежность задачи
	if task.GetUserId() != userID {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTask - обновление задачи
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}
	}
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Проверяем принадлежность задачи
	task, err := h.taskClient.GetTask(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	
	if task.GetUserId() != userID {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var input struct {
		Text   string `json:"text"`
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	updatedTask, err := h.taskClient.UpdateTask(r.Context(), int32(id), input.Text, input.Status)
	if err != nil {
		log.Printf("UpdateTask error: %v", err)
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}

// DeleteTask - удаление задачи
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}
	}
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Проверяем принадлежность задачи
	task, err := h.taskClient.GetTask(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	
	if task.GetUserId() != userID {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	success, err := h.taskClient.DeleteTask(r.Context(), int32(id))
	if err != nil {
		log.Printf("DeleteTask error: %v", err)
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}
	
	if !success {
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "task deleted",
		"id":      id,
	})
}