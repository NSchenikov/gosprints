package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "api-gateway/internal/grpc/client"  // gRPC клиент к task-service
    "api-gateway/pkg/auth"
)

type TaskProxyHandler struct {
    taskClient *client.TaskClient  // gRPC клиент к task-service
}

func (h *TaskProxyHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
    //  берем user_id из JWT
    userID, err := auth.GetUserFromJWT(r)
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Проксирование запроса в task-service через gRPC
    tasks, total, err := h.taskClient.ListTasks(r.Context(), userID, 
        r.URL.Query().Get("status"), 
        parseInt(r.URL.Query().Get("page"), 1),
        parseInt(r.URL.Query().Get("page_size"), 10))
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    //Отправляем клиенту ответ 
    json.NewEncoder(w).Encode(map[string]interface{}{
        "tasks": tasks,
        "total": total,
    })
}