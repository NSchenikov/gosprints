
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"log"
	"time"

	"gosprints/internal/grpc/task/client"
	"gosprints/internal/models"
	"gosprints/pkg/auth"
	// "gosprints/internal/ws"
	"gosprints/internal/kafka"
)

type TaskHandler struct {
	taskClient *client.TaskClient
	// hub        *ws.NotificationHub
	eventProd  *kafka.TaskEventProducer
}

func NewTaskHandler(taskClient *client.TaskClient, hub *ws.NotificationHub) *TaskHandler {
	return &TaskHandler{
		taskClient: taskClient,
		// hub:        hub,
		eventProd:  eventProd,
	}
}
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {

	userID, err := auth.GetUserFromJWT(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	// берем параметры пагинации из query
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	
	status := r.URL.Query().Get("status")
	
	// вызов gRPC-клиента
	tasks, total, err := h.taskClient.ListTasks(r.Context(), userID, status, int32(page), int32(pageSize))
	if err != nil {
		log.Printf("Error getting tasks: %v\n", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	
	// Конвертация proto задач в модели
	var responseTasks []models.Task
	for _, protoTask := range tasks {
		responseTasks = append(responseTasks, models.Task{
			ID:        int(protoTask.GetId()),
			Text:      protoTask.GetText(),
			Status:    protoTask.GetStatus(),
			UserID:    protoTask.GetUserId(),
			CreatedAt: protoTask.GetCreatedAt().AsTime(),
		})
	}
	
	//инфо о пагинации
	response := map[string]interface{}{
		"tasks":     responseTasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Println("All tasks response sent")
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {

	userID, err := auth.GetUserFromJWT(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	//gRPC клиент
	task, err := h.taskClient.GetTask(r.Context(), int32(id))
	if err != nil {
		log.Printf("Error getting task: %v\n", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	
	// задача принадлежит пользователю?
	if task.GetUserId() != userID {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Конвертация в модель
	response := models.Task{
		ID:        int(task.GetId()),
		Text:      task.GetText(),
		Status:    task.GetStatus(),
		UserID:    task.GetUserId(),
		CreatedAt: task.GetCreatedAt().AsTime(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Println("Task response sent")
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserFromJWT(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	//gRPC клиент
	task, err := h.taskClient.CreateTask(r.Context(), input.Text, userID)
	if err != nil {
		log.Printf("CreateTask error: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	if h.eventProd != nil {
        go h.eventProd.PublishTaskEvent(r.Context(), "CREATED", 
            task.GetId(), task.GetText(), task.GetStatus(), userID)
    }

	// публикуется событие в Kafka
	if h.eventProducer != nil {
		event := &events.TaskEvent{
			EventId:    uuid.New().String(),
			EventType:  "CREATED",
			TaskId:     int32(task.GetId()),
			TaskText:   task.GetText(),
			TaskStatus: task.GetStatus(),
			UserId:     userID,
			Timestamp:  timestamppb.Now(),
		}
		h.eventProducer.Publish(ctx, event)
	}

	// Конвертируем в модель для ответа
	response := models.Task{
		ID:        int(task.GetId()),
		Text:      task.GetText(),
		Status:    task.GetStatus(),
		UserID:    task.GetUserId(),
		CreatedAt: task.GetCreatedAt().AsTime(),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserFromJWT(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// задача принадлежит пользователю?
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

	// gRPC клиент
	updatedTask, err := h.taskClient.UpdateTask(r.Context(), int32(id), input.Text, input.Status)
	if err != nil {
		log.Printf("UpdateTask error: %v", err)
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	// публикация события в Kafka
	if h.eventProducer != nil {
    event := &events.TaskEvent{
        EventId:    uuid.New().String(),
        EventType:  "CREATED",
        TaskId:     int32(task.GetId()),
        TaskText:   task.GetText(),
        TaskStatus: task.GetStatus(),
        UserId:     userID,
        Timestamp:  timestamppb.Now(),
    }
    h.eventProducer.Publish(ctx, event)
}

	// Конвертация в модель
	response := models.Task{
		ID:        int(updatedTask.GetId()),
		Text:      updatedTask.GetText(),
		Status:    updatedTask.GetStatus(),
		UserID:    updatedTask.GetUserId(),
		CreatedAt: updatedTask.GetCreatedAt().AsTime(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserFromJWT(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// задача принадлежит пользователю?
	task, err := h.taskClient.GetTask(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	
	if task.GetUserId() != userID {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	//gRPC клиент
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

	// Публикуем событие в Kafka
	if h.eventProducer != nil {
		event := &events.TaskEvent{
			EventId:    uuid.New().String(),
			EventType:  "CREATED",
			TaskId:     int32(task.GetId()),
			TaskText:   task.GetText(),
			TaskStatus: task.GetStatus(),
			UserId:     userID,
			Timestamp:  timestamppb.Now(),
		}
		h.eventProducer.Publish(ctx, event)
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "task deleted",
		"id":      id,
	})
}

func (h *TaskHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
    userID, err := auth.GetUserFromJWT(r)
    if err != nil {
        // log.Printf("[SearchTasks] Auth error: %v", err)
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    // log.Printf("[SearchTasks] UserID: %s", userID)
    
    query := r.URL.Query().Get("q")
    if query == "" {
        // log.Printf("[SearchTasks] Empty query")
        http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
        return
    }
    // log.Printf("[SearchTasks] Query: %s", query)
    
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    
    pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
    if pageSize < 1 {
        pageSize = 10
    }
    // log.Printf("[SearchTasks] Page: %d, PageSize: %d", page, pageSize)
    
    tasks, total, err := h.taskClient.SearchTasks(r.Context(), query, userID, int32(page), int32(pageSize))
    if err != nil {
        // log.Printf("[SearchTasks] gRPC error: %v", err)
        http.Error(w, "Search failed", http.StatusInternalServerError)
        return
    }
    // log.Printf("[SearchTasks] Found %d tasks, total: %d", len(tasks), total)
    
    // Конвертируем proto задачи в модели
    var responseTasks []models.Task
    for _, protoTask := range tasks {
        responseTasks = append(responseTasks, models.Task{
            ID:        int(protoTask.GetId()),
            Text:      protoTask.GetText(),
            Status:    protoTask.GetStatus(),
            UserID:    protoTask.GetUserId(),
            CreatedAt: protoTask.GetCreatedAt().AsTime(),
        })
    }
    
    response := map[string]interface{}{
        "tasks":     responseTasks,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
    // log.Printf("[SearchTasks] Response sent")
}