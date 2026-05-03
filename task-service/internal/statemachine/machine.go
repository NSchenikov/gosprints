package statemachine

import (
    "context"
    "log"
    "strings"
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
    // 1. NEW → VALIDATION_1
    newTasks, _ := sm.repo.GetByStatus(ctx, models.TaskStatusNew)
    for _, task := range newTasks {
        if sm.performValidation1(&task) {
            sm.repo.UpdateStatus(ctx, task.ID, models.TaskStatusValidation1, nil, nil)
            log.Printf("[StateMachine] Task %d: passed VALIDATION_1", task.ID)
        } else {
            sm.repo.UpdateStatus(ctx, task.ID, models.TaskStatusFailed, nil, nil)
            log.Printf("[StateMachine] Task %d: failed VALIDATION_1", task.ID)
        }
    }
    
    // 2. VALIDATION_1 → WAITING_FOR_VALIDATION_2
    validation1Tasks, _ := sm.repo.GetByStatus(ctx, models.TaskStatusValidation1)
    for _, task := range validation1Tasks {
        sm.repo.UpdateStatus(ctx, task.ID, models.TaskStatusWaitingForValidation2, nil, nil)
        log.Printf("[StateMachine] Task %d: → WAITING_FOR_VALIDATION_2", task.ID)
    }
    
    // 3. Обработка WAITING_FOR_VALIDATION_2
    waitingTasks, _ := sm.repo.GetByStatus(ctx, models.TaskStatusWaitingForValidation2)
    for _, task := range waitingTasks {
        if time.Since(task.CreatedAt) > 60*time.Minute {
            sm.repo.Delete(ctx, task.ID)
            log.Printf("[StateMachine] Task %d: expired, deleted", task.ID)
            continue
        }
        
        if task.Attempts >= 3 {
            sm.repo.UpdateStatus(ctx, task.ID, models.TaskStatusFailed, nil, nil)
            sm.sendNotification(ctx, &task, "TASK_FAILED", "Task failed after max attempts")
            log.Printf("[StateMachine] Task %d: failed after %d attempts", task.ID, task.Attempts)
            continue
        }
        
        if sm.performValidation2(ctx, &task) {
            sm.repo.UpdateStatus(ctx, task.ID, models.TaskStatusReadyForClosure, nil, nil)
            sm.sendNotification(ctx, &task, "TASK_READY", "Task ready for closure")
            log.Printf("[StateMachine] Task %d: passed VALIDATION_2 → READY_FOR_CLOSURE", task.ID)
        } else {
            sm.repo.IncrementAttempts(ctx, task.ID)
            log.Printf("[StateMachine] Task %d: failed VALIDATION_2, attempt %d/3", task.ID, task.Attempts+1)
        }
    }
    
    // 4. READY_FOR_CLOSURE → CLOSED (автозакрытие через 1 час)
    readyTasks, _ := sm.repo.GetByStatus(ctx, models.TaskStatusReadyForClosure)
    for _, task := range readyTasks {
        if time.Since(task.UpdatedAt) > 1*time.Hour {
            sm.repo.UpdateStatus(ctx, task.ID, models.TaskStatusClosed, nil, nil)
            sm.sendNotification(ctx, &task, "TASK_CLOSED", "Task auto-closed")
            log.Printf("[StateMachine] Task %d: auto-closed", task.ID)
        }
    }
}

func (sm *TaskStateMachine) performValidation1(task *models.Task) bool {
    // Проверка: текст не пустой
    if strings.TrimSpace(task.Text) == "" {
        log.Printf("[StateMachine] Task %d: validation1 failed - empty text", task.ID)
        return false
    }
    // Проверка: текст не слишком длинный (максимум 1000 символов)
    if len(task.Text) > 1000 {
        log.Printf("[StateMachine] Task %d: validation1 failed - text too long (%d chars)", task.ID, len(task.Text))
        return false
    }
    return true
}

func (sm *TaskStateMachine) performValidation2(ctx context.Context, task *models.Task) bool {
    // Проверка лимита активных задач пользователя (максимум 5)
    activeCount, err := sm.repo.GetActiveTasksCount(ctx, task.UserID)
    if err != nil {
        log.Printf("[StateMachine] Task %d: validation2 error - %v", task.ID, err)
        return false
    }
    if activeCount >= 5 {
        log.Printf("[StateMachine] Task %d: validation2 failed - user has %d active tasks (max 5)", task.ID, activeCount)
        return false
    }
    
    //  TODO: Проверка, есть ли свободные воркеры
    // (можно добавить проверку очереди или количества активных воркеров)
    // Для демо: всегда true
    return true
}

func (sm *TaskStateMachine) sendNotification(ctx context.Context, task *models.Task, eventType, message string) {
    if sm.producer != nil {
        go sm.producer.PublishTaskEvent(ctx, eventType, 
            int32(task.ID), task.Text, task.Status, task.UserID)
    }
}