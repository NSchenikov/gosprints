package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gosprints/internal/models"
	"gosprints/internal/services"
)

type TaskHandler struct {
	service services.TaskService
}

func NewTaskHandler(service services.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tasks, err := h.service.GetTasks(ctx)
	if err != nil {
		fmt.Printf("Error getting tasks: %v\n", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tasks)
	fmt.Println("All tasks response sent")
}

// посмотреть задачи
func (h *TaskHandler) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	str, _ := h.service.GetTasks(r.Context())
	json.NewEncoder(w).Encode(str)
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.service.GetTaskByID(r.Context(), id)
	if err != nil {
		fmt.Printf("Error getting task: %v\n", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
	fmt.Println("Task response sent")
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	created, err := h.service.CreateTask(r.Context(), &task)
	if err != nil {
		http.Error(w, "Failed to insert into DB", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

	updated, err := h.service.UpdateTask(r.Context(), id, &task)
	if err != nil {
		http.Error(w, "failed to update task", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
        idStr := r.URL.Path[len("/tasks/"):]
		id, err := strconv.Atoi(idStr)
        if err != nil {
            http.Error(w, "Task ID is required", http.StatusBadRequest)
            return
        }

        if err := h.service.DeleteTask(r.Context(), id); err != nil {
			http.Error(w, "failed to delete task", http.StatusInternalServerError)
			return
        }
        
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    _ = json.NewEncoder(w).Encode(map[string]any{
        "message": "task deleted",
        "id":      id,
    })
}