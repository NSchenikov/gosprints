package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gosprints/internal/models"
	"gosprints/internal/repositories"
)

type TaskHandler struct {
	taskRepo repositories.TaskRepository
}

func NewTaskHandler(taskRepo repositories.TaskRepository) *TaskHandler {
	return &TaskHandler{taskRepo: taskRepo}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	
	tasks, err := h.taskRepo.GetAll()
	if err != nil {
		fmt.Printf("Error getting tasks: %v\n", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Retrieved %d tasks from database\n", len(tasks))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
	fmt.Println("All tasks response sent")
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.taskRepo.GetByID(id)
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
	
	var newTask struct {
		Text string `json:"text"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if newTask.Text == "" {
		http.Error(w, "Text field is required", http.StatusBadRequest)
		return
	}
	
	task := &models.Task{Text: newTask.Text}
	err := h.taskRepo.Create(task)
	if err != nil {
		fmt.Printf("Error creating task: %v\n", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
	fmt.Printf("Task created: ID=%d\n", task.ID)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var updatedTask struct {
		Text string `json:"text"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if updatedTask.Text == "" {
		http.Error(w, "Text field is required", http.StatusBadRequest)
		return
	}

	err = h.taskRepo.Update(id, updatedTask.Text)
	if err != nil {
		fmt.Printf("Error updating task: %v\n", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Task updated successfully",
	})
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
        idStr := r.URL.Path[len("/tasks/"):]
        if idStr == "" {
            http.Error(w, "Task ID is required", http.StatusBadRequest)
            return
        }

        var id int
        if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
            http.Error(w, "Invalid task ID", http.StatusBadRequest)
            return
        }

        task, err := h.taskRepo.GetByID(id)
        if err != nil {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "error": err.Error(),
            })
            return
        }

        err = h.taskRepo.Delete(id)
        if err != nil {
            fmt.Printf("Error deleting task: %v\n", err)
            if strings.Contains(err.Error(), "not found") {
                http.Error(w, "Task not found", http.StatusNotFound)
            } else {
                http.Error(w, "Database error", http.StatusInternalServerError)
            }
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        if err := json.NewEncoder(w).Encode(map[string]interface{}{
            "message": "Task deleted successfully",
            "deleted_task": task,
        }); err != nil {
            fmt.Printf("JSON encoding error: %v\n", err)
            return
        }
}