package services

import (
    "context"
    "time"

    "gosprints/internal/models"
)

type TaskRepository interface {
	GetAll(ctx context.Context) ([]models.Task, error)
	GetByID(ctx context.Context, id int) (*models.Task, error)
	Create(ctx context.Context, task *models.Task) (int, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id int) error
    UpdateStatus(ctx context.Context, id int, status string, startedAt, endedAt *time.Time) error
    GetByStatus(ctx context.Context, status string) ([]models.Task, error)
}

type TaskCacheRepository interface {
    TaskRepository
    WarmUpCache(ctx context.Context) error
    ClearCache(ctx context.Context) error
    GetCacheStats() CacheStats
}

type CacheStats struct {
    Hits       int64
    Misses     int64
    Sets       int64
    Deletes    int64
    Expirations int64
}

type taskService struct {
    repo TaskRepository
}

func NewTaskService(repo TaskRepository) *taskService {
    return &taskService{repo: repo}
}

func (s *taskService) GetTasks(ctx context.Context) ([]models.Task, error) {
    return s.repo.GetAll(ctx)
}

func (s *taskService) GetTaskByID(ctx context.Context, id int) (models.Task, error) {
    t, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return models.Task{}, err
    }
    return *t, nil
}

func (s *taskService) CreateTask(ctx context.Context, task *models.Task) (models.Task, error) {
    if task.Status == "" {
        task.Status = "pending"
    }

    id, err := s.repo.Create(ctx, task)
    if err != nil {
        return models.Task{}, err
    }
    task.ID = id

    return *task, nil
}

func (s *taskService) UpdateTask(ctx context.Context, id int, task *models.Task) (models.Task, error) {
    task.ID = id

    if err := s.repo.Update(ctx, task); err != nil {
        return models.Task{}, err
    }

    updated, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return models.Task{}, err
    }

    return *updated, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id int) error {
    return s.repo.Delete(ctx, id)
}
