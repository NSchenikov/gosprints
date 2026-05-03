package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    "log"
    
    "api-gateway/internal/grpc/client"
)

type TaskProxyHandler struct {
    taskClient *client.TaskClient
}

func NewTaskProxyHandler(taskClient *client.TaskClient) *TaskProxyHandler {
    return &TaskProxyHandler{taskClient: taskClient}
}

func parseInt(s string, defaultVal int) int {
    if s == "" {
        return defaultVal
    }
    val, err := strconv.Atoi(s)
    if err != nil {
        return defaultVal
    }
    return val
}

func (h *TaskProxyHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    page := parseInt(r.URL.Query().Get("page"), 1)
    pageSize := parseInt(r.URL.Query().Get("page_size"), 10)
    status := r.URL.Query().Get("status")
    
    tasks, total, err := h.taskClient.ListTasks(r.Context(), userID, status, int32(page), int32(pageSize))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
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

func (h *TaskProxyHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    var req struct {
        Text string `json:"text"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    task, err := h.taskClient.CreateTaskWithStatus(r.Context(), req.Text, userID, "NEW")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}

func (h *TaskProxyHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    idStr := r.URL.Path[len("/tasks/"):]
    id := parseInt(idStr, 0)
    if id == 0 {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    
    task, err := h.taskClient.GetTask(r.Context(), int32(id))
    if err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    if task.GetUserId() != userID {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}

func (h *TaskProxyHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    idStr := r.URL.Path[len("/tasks/"):]
    id := parseInt(idStr, 0)
    if id == 0 {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    
    // Проверяем принадлежность
    task, err := h.taskClient.GetTask(r.Context(), int32(id))
    if err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    if task.GetUserId() != userID {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    var req struct {
        Text   string `json:"text"`
        Status string `json:"status"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    updatedTask, err := h.taskClient.UpdateTask(r.Context(), int32(id), req.Text, req.Status)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updatedTask)
}

func (h *TaskProxyHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    idStr := r.URL.Path[len("/tasks/"):]
    id := parseInt(idStr, 0)
    if id == 0 {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    
    // Проверяем принадлежность
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
    if err != nil || !success {
        http.Error(w, "Failed to delete task", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "task deleted",
        "id":      id,
    })
}

func (h *TaskProxyHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    query := r.URL.Query().Get("q")
    if query == "" {
        http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
        return
    }
    
    page := parseInt(r.URL.Query().Get("page"), 1)
    pageSize := parseInt(r.URL.Query().Get("page_size"), 10)
    
    tasks, total, err := h.taskClient.SearchTasks(r.Context(), query, userID, int32(page), int32(pageSize))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
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

// CloseTask - закрытие задачи
func (h *TaskProxyHandler) CloseTask(w http.ResponseWriter, r *http.Request) {
    log.Println("[CloseTask] Called")
    
    userID := r.Context().Value("user_id").(string)
    log.Printf("[CloseTask] UserID: %s", userID)
    
    // Извлекаем ID из URL /tasks/{id}/close
    idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
    idStr = strings.TrimSuffix(idStr, "/close")
    log.Printf("[CloseTask] ID string: %s", idStr)
    
    id, err := strconv.Atoi(idStr)
    if err != nil {
        log.Printf("[CloseTask] Invalid ID: %v", err)
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    log.Printf("[CloseTask] Task ID: %d", id)
    
    // Получаем задачу
    task, err := h.taskClient.GetTask(r.Context(), int32(id))
    if err != nil {
        log.Printf("[CloseTask] GetTask error: %v", err)
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    log.Printf("[CloseTask] Task status: %s, UserID: %s", task.GetStatus(), task.GetUserId())
    
    if task.GetUserId() != userID {
        log.Printf("[CloseTask] Forbidden: user %s != %s", userID, task.GetUserId())
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    if task.GetStatus() != "READY_FOR_CLOSURE" {
        log.Printf("[CloseTask] Wrong status: %s", task.GetStatus())
        http.Error(w, "Task not ready for closure", http.StatusBadRequest)
        return
    }
    
    updatedTask, err := h.taskClient.UpdateTask(r.Context(), int32(id), task.GetText(), "CLOSED")
    if err != nil {
        log.Printf("[CloseTask] UpdateTask error: %v", err)
        http.Error(w, "Failed to close task", http.StatusInternalServerError)
        return
    }
    
    log.Printf("[CloseTask] Task %d closed successfully", id)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id":     updatedTask.GetId(),
        "status": updatedTask.GetStatus(),
        "closed": true,
    })
}

// GetUserTasks - список задач пользователя
func (h *TaskProxyHandler) GetUserTasks(w http.ResponseWriter, r *http.Request) {
    // Извлекаем user_id из URL
    userID := strings.TrimPrefix(r.URL.Path, "/users/")
    userID = strings.TrimSuffix(userID, "/tasks")
    
    if userID == "" {
        http.Error(w, "user_id is required", http.StatusBadRequest)
        return
    }
    
    tasks, total, err := h.taskClient.ListTasks(r.Context(), userID, "", 1, 100)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "tasks": tasks,
        "total": total,
    })
}