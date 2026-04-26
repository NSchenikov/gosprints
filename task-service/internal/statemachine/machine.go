package statemachine

import (
    "context"
    "log"
    "time"
    
    "task-service/internal/models"
    "task-service/internal/repositories"
    "task-service/internal/kafka"
)

type TaskStateMachine struct {
    repo     repositories.TaskRepository
    producer *kafka.TaskEventProducer
    ticker   *time.Ticker
    stopCh   chan struct{}
}

func NewTaskStateMachine(repo repositories.TaskRepository, producer *kafka.TaskEventProducer) *TaskStateMachine {
    return &TaskStateMachine{
        repo:     repo,
        producer: producer,
        ticker:   time.NewTicker(30 * time.Second),
        stopCh:   make(chan struct{}),
    }
}

func (sm *TaskStateMachine) Start(ctx context.Context) {
    log.Println("[StateMachine] Started")
    
    for {
        select {
        case <-sm.ticker.C:
            sm.processTasks(ctx)
        case <-sm.stopCh:
            log.Println("[StateMachine] Stopped")
            return
        case <-ctx.Done():
            return
        }
    }
}

func (sm *TaskStateMachine) Stop() {
    sm.ticker.Stop()
    close(sm.stopCh)
}

func (sm *TaskStateMachine) processTasks(ctx context.Context) {
    // Обработка задачи в WAITING_FOR_VALIDATION_2
    waitingTasks, err := sm.repo.GetByStatus(ctx, models.TaskStatusWaitingForValidation2)
    if err != nil {
        log.Printf("[StateMachine] Error fetching waiting tasks: %v", err)
        return
    }
    
    for _, task := range waitingTasks {
        // Проверяем, не зависла ли задача дольше чем на 60 минут
        if time.Since(task.CreatedAt) > 60*time.Minute {
            log.Printf("[StateMachine] Task %d expired, deleting", task.ID)
            sm.repo.Delete(ctx, task.ID)
            continue
        }
        
        // Проверяем попытки attempts
        if task.Attempts >= 3 {
            log.Printf("[StateMachine] Task %d failed after %d attempts", task.ID, task.Attempts)
            sm.repo.UpdateStatus(ctx, task.ID, models.TaskStatusFailed, nil, nil)
            sm.sendNotification(ctx, &task, "TASK_FAILED", "Task failed after max attempts")
            continue
        }
        
        // Пробуем валидацию
        success := sm.performValidation2(ctx, &task)
        
        if success {
            log.Printf("[StateMachine] Task %d passed VALIDATION_2", task.ID)
            sm.repo.UpdateStatus(ctx, task.ID, models.TaskStatusReadyForClosure, nil, nil)
            sm.sendNotification(ctx, &task, "TASK_READY", "Task ready for closure")
        } else {
            sm.repo.IncrementAttempts(ctx, task.ID)
            log.Printf("[StateMachine] Task %d validation failed, attempt %d/%d", 
                task.ID, task.Attempts+1, 3)
        }
    }
    
    // Автозакрытие задач в READY_FOR_CLOSURE
    readyTasks, err := sm.repo.GetByStatus(ctx, models.TaskStatusReadyForClosure)
    if err != nil {
        log.Printf("[StateMachine] Error fetching ready tasks: %v", err)
        return
    }
    
    for _, task := range readyTasks {
        if time.Since(task.UpdatedAt) > 1*time.Hour {
            log.Printf("[StateMachine] Auto-closing task %d", task.ID)
            sm.repo.UpdateStatus(ctx, task.ID, models.TaskStatusClosed, nil, nil)
            sm.sendNotification(ctx, &task, "TASK_CLOSED", "Task auto-closed")
        }
    }
}

func (sm *TaskStateMachine) performValidation2(ctx context.Context, task *models.Task) bool {
    // TODO: Валидация, есть ли свободные воркеры?
    // Можно проверить очередь или количество активных воркеров
    // Демо вариант: успех с вероятностью 70%
    return time.Now().UnixNano()%10 < 7 //
}

func (sm *TaskStateMachine) sendNotification(ctx context.Context, task *models.Task, eventType, message string) {
    if sm.producer != nil {
        go sm.producer.PublishTaskEvent(ctx, eventType, 
            int32(task.ID), task.Text, task.Status, task.UserID)
    }
}