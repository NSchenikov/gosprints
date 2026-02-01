package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"context"
	"time"
	"log"

	"gosprints/internal/models"
	"gosprints/pkg/auth"
	"gosprints/internal/grpc/task/client"
	"gosprints/internal/grpc/task/pb"
	"gosprints/internal/services"
	"gosprints/internal/ws"
)

type TaskService interface {
    GetTasks(ctx context.Context) ([]models.Task, error)
    GetTaskByID(ctx context.Context, id int) (models.Task, error)
    CreateTask(ctx context.Context, task *models.Task) (models.Task, error)
    UpdateTask(ctx context.Context, id int, task *models.Task) (models.Task, error)
    DeleteTask(ctx context.Context, id int) error
}

type TaskHandler struct {
	service TaskService
	taskClient *client.TaskClient     // новый gRPC клиент
    hub        *ws.NotificationHub
}

func NewTaskHandler(service TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
		hub: nil, //возможно придется создать
	}
}

//!конструктор для работы с gRPC
func NewTaskHandlerWithGRPC(taskClient *client.TaskClient, hub *ws.NotificationHub) *TaskHandler {
    return &TaskHandler{
        taskClient: taskClient,
        hub:        hub,
        service:    nil, // не используем прямой сервис
    }
}

//комбинированный конструктор
func NewTaskHandlerWithBoth(service TaskService, taskClient *client.TaskClient, hub *ws.NotificationHub) *TaskHandler {
	return &TaskHandler{
		service:    service,
		taskClient: taskClient,
		hub:        hub,
	}
}

//преобразование gRPC Task в models.Task
func convertProtoToModel(protoTask *pb.Task) models.Task {
	task := models.Task{
		ID:        protoTask.GetId(),
		Text:      protoTask.GetText(),
		Status:    protoTask.GetStatus(),
		UserID:    protoTask.GetUserId(),
		CreatedAt: protoTask.GetCreatedAt().AsTime(),
	}

	// Обрабатываем optional поля
	if protoTask.GetStartedAt() != nil {
		startedAt := protoTask.GetStartedAt().AsTime()
		task.StartedAt = &startedAt
	}

	if protoTask.GetEndedAt() != nil {
		endedAt := protoTask.GetEndedAt().AsTime()
		task.EndedAt = &endedAt
	}

	return task
}

// извлечение ID из URL
func extractID(path, prefix string) (string, error) {
	idStr := strings.TrimPrefix(path, prefix)
	if idStr == "" {
		return "", fmt.Errorf("task ID is required")
	}
	return idStr, nil
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Используем gRPC если доступен
	if h.taskClient != nil {
		tasks, total, err := h.taskClient.ListTasks(ctx, "", "", 1, 100)
		if err != nil {
			log.Printf("gRPC GetTasks error: %v", err)
			http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
			return
		}

		// Конвертируем в модели
		var modelTasks []models.Task
		for _, protoTask := range tasks {
			modelTasks = append(modelTasks, convertProtoToModel(protoTask))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modelTasks)
		return
	}

	// возвращаем кусочек старой реализации
	if h.service != nil {
		tasks, err := h.service.GetTasks(ctx)
		if err != nil {
			fmt.Printf("Error getting tasks: %v\n", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
		return
	}

	http.Error(w, "Handler not configured", http.StatusInternalServerError)
}

// посмотреть задачи
func (h *TaskHandler) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	if h.service != nil {
		str, _ := h.service.GetTasks(r.Context())
		json.NewEncoder(w).Encode(str)
		return
	}
	http.Error(w, "Handler not configured", http.StatusInternalServerError)
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Извлекаем ID из URL
	idStr, err := extractID(r.URL.Path, "/tasks/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Используем gRPC если доступен
	if h.taskClient != nil {
		task, err := h.taskClient.GetTask(ctx, idStr)
		if err != nil {
			log.Printf("gRPC GetTask error: %v", err)
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		modelTask := convertProtoToModel(task)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modelTask)
		return
	}

	// Fallback к старой реализации
	if h.service != nil {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		task, err := h.service.GetTaskByID(ctx, id)
		if err != nil {
			fmt.Printf("Error getting task: %v\n", err)
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
		return
	}

	http.Error(w, "Handler not configured", http.StatusInternalServerError)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	// Используем gRPC если доступен
	if h.taskClient != nil {
		task, err := h.taskClient.CreateTask(ctx, input.Text, userID)
		if err != nil {
			log.Printf("gRPC CreateTask error: %v", err)
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		// уведомление через WebSocket
		if h.hub != nil {
			event := models.TaskStatusEvent{
				TaskID:    task.GetId(),
				Status:    task.GetStatus(),
				UserID:    userID,
				Timestamp: time.Now(),
			}
			h.hub.SendToUser(userID, event)
		}

		// Конвертация ответа
		modelTask := convertProtoToModel(task)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(modelTask)
		return
	}

	// Fallback к старой реализации
	if h.service != nil {
		task := &models.Task{
			Text:   input.Text,
			UserID: userID,
		}

		created, err := h.service.CreateTask(ctx, task)
		if err != nil {
			log.Printf("CreateTask DB error: %v", err)
			http.Error(w, "Failed to insert into DB", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
		return
	}

	http.Error(w, "Handler not configured", http.StatusInternalServerError)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Извлекаем ID из URL
	idStr, err := extractID(r.URL.Path, "/tasks/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var input struct {
		Text   string `json:"text"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Используем gRPC если доступен
	if h.taskClient != nil {
		task, err := h.taskClient.UpdateTask(ctx, idStr, input.Text, input.Status)
		if err != nil {
			log.Printf("gRPC UpdateTask error: %v", err)
			http.Error(w, "Failed to update task", http.StatusNotFound)
			return
		}

		modelTask := convertProtoToModel(task)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modelTask)
		return
	}

	// Fallback к старой реализации
	if h.service != nil {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		task := &models.Task{
			Text:   input.Text,
			Status: input.Status,
		}

		updated, err := h.service.UpdateTask(ctx, id, task)
		if err != nil {
			http.Error(w, "Failed to update task", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
		return
	}

	http.Error(w, "Handler not configured", http.StatusInternalServerError)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Извлекаем ID из URL
	idStr, err := extractID(r.URL.Path, "/tasks/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Используем gRPC если доступен
	if h.taskClient != nil {
		success, err := h.taskClient.DeleteTask(ctx, idStr)
		if err != nil {
			log.Printf("gRPC DeleteTask error: %v", err)
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			return
		}

		if !success {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "task deleted",
			"id":      idStr,
		})
		return
	}

	// Fallback к старой реализации
	if h.service != nil {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		if err := h.service.DeleteTask(ctx, id); err != nil {
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "task deleted",
			"id":      id,
		})
		return
	}

	http.Error(w, "Handler not configured", http.StatusInternalServerError)
}