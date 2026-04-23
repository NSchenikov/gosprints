package models

import "time"

type TaskEvent struct {
    EventId     string    `json:"event_id"`
    EventType   string    `json:"event_type"`   // CREATED, UPDATED, COMPLETED
    TaskId      int32     `json:"task_id"`
    TaskText    string    `json:"task_text"`
    TaskStatus  string    `json:"task_status"`
    UserId      string    `json:"user_id"`
    Timestamp   time.Time `json:"timestamp"`
}

// Аналитические данные
type TaskAnalytics struct {
    UserId            string    `json:"user_id"`
    // TasksCreated      int32     `json:"tasks_created"`
    TasksCompleted    int32     `json:"tasks_completed"`
    AvgCompletionTime float64   `json:"avg_completion_time"`
    LastEventTime     time.Time `json:"last_event_time"`
    Date              time.Time `json:"date"` // для агрегации по дням
}