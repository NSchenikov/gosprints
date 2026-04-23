package processor

import (
    "context"
    "log"
    "sync"
    "time"
    
    events "etl-worker/proto/events"
    "etl-worker/internal/models"
    "etl-worker/internal/storage"
)

type EventProcessor struct {
    storage   storage.AnalyticsStorage
    userStats sync.Map // userId -> *models.TaskAnalytics
    taskStartTime sync.Map // taskId -> time.Time
}

func NewEventProcessor(storage storage.AnalyticsStorage) *EventProcessor {
    return &EventProcessor{
        storage: storage,
    }
}

func (p *EventProcessor) ProcessEvent(ctx context.Context, event *events.TaskEvent) error {
    log.Printf("[ETL] Processing event: type=%s, taskID=%d", event.EventType, event.TaskId)
    
    // Обновляем статистику в зависимости от типа события
    switch event.EventType {
    case "CREATED":
        // Сохраняем время создания задачи
        p.taskStartTime.Store(event.TaskId, event.Timestamp.AsTime())
        log.Printf("[ETL] Stored start time for task %d", event.TaskId)
        
    case "COMPLETED":
        // Получаем время создания
        startTimeVal, ok := p.taskStartTime.Load(event.TaskId)
        if !ok {
            log.Printf("[ETL] No start time found for task %d, skipping", event.TaskId)
            return nil
        }
        
        startTime := startTimeVal.(time.Time)
        endTime := event.Timestamp.AsTime()
        duration := endTime.Sub(startTime).Seconds()
        
        log.Printf("[ETL] Task %d completed in %.2f seconds", event.TaskId, duration)
        
        // Обновляем статистику пользователя
        stats := p.getOrCreateUserStats(event.UserId)
        oldAvg := stats.AvgCompletionTime
        oldCount := stats.TasksCompleted
        
        stats.TasksCompleted++
        stats.AvgCompletionTime = (oldAvg*float64(oldCount) + duration) / float64(stats.TasksCompleted)
        stats.LastEventTime = endTime
        stats.Date = endTime.Truncate(24 * time.Hour)
        
        log.Printf("[ETL] User %s: completed=%d, avg=%.2f", 
            event.UserId, stats.TasksCompleted, stats.AvgCompletionTime)
        
        // Сохраняем в ClickHouse
        if err := p.storage.SaveAnalytics(ctx, stats); err != nil {
            log.Printf("[ETL] Failed to save analytics: %v", err)
            return err
        }
        
        p.userStats.Store(event.UserId, stats)
        p.taskStartTime.Delete(event.TaskId)
        
    case "UPDATED":
        log.Printf("[ETL] Task %d updated, status: %s", event.TaskId, event.TaskStatus)
    }
    
    return nil
}

func (p *EventProcessor) getOrCreateUserStats(userID string) *models.TaskAnalytics {
    if val, ok := p.userStats.Load(userID); ok {
        return val.(*models.TaskAnalytics)
    }
    
    stats := &models.TaskAnalytics{
        UserId:        userID,
        TasksCompleted: 0,
        AvgCompletionTime: 0,
        LastEventTime: time.Now(),
        Date:          time.Now().Truncate(24 * time.Hour),
    }
    
    p.userStats.Store(userID, stats)
    return stats
}