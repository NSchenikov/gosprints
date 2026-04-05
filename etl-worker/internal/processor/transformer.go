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
    userStats sync.Map // map[string]*models.TaskAnalytics
}

func NewEventProcessor(storage storage.AnalyticsStorage) *EventProcessor {
    return &EventProcessor{
        storage: storage,
    }
}

func (p *EventProcessor) ProcessEvent(ctx context.Context, event *events.TaskEvent) error {
    log.Printf("[ETL] Processing event: type=%s, taskID=%d", event.EventType, event.TaskId)
    
    // Получаем или создаём статистику для пользователя
    stats := p.getOrCreateUserStats(event.UserId)
    
    // Обновляем статистику в зависимости от типа события
    switch event.EventType {
    case "CREATED":
        stats.TasksCreated++
        stats.LastEventTime = event.Timestamp.AsTime()
        
    case "COMPLETED":
        stats.TasksCompleted++
        stats.LastEventTime = event.Timestamp.AsTime()
        
        // TODO: вычисления время выполнения задачи
        // нужно положить время создания задачи в отдельное хранилище
        
    case "UPDATED":
        stats.LastEventTime = event.Timestamp.AsTime()
    }
    
    // Укладываем статистику в хранилище
    if err := p.storage.SaveAnalytics(ctx, stats); err != nil {
        log.Printf("[ETL] Failed to save analytics: %v", err)
        return err
    }
    
    // Обновление кэша
    p.userStats.Store(event.UserId, stats)
    
    return nil
}

func (p *EventProcessor) getOrCreateUserStats(userID string) *models.TaskAnalytics {
    if val, ok := p.userStats.Load(userID); ok {
        return val.(*models.TaskAnalytics)
    }
    
    stats := &models.TaskAnalytics{
        UserId:        userID,
        TasksCreated:  0,
        TasksCompleted: 0,
        LastEventTime: time.Now(),
        Date:          time.Now().Truncate(24 * time.Hour),
    }
    
    p.userStats.Store(userID, stats)
    return stats
}