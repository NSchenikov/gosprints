package services

import (
    "context"

    "gosprints/internal/models"
    "gosprints/internal/repositories"
)

type TaskService interface {
    GetTasks(ctx context.Context) ([]models.Task, error)
    GetTaskByID(ctx context.Context, id int) (models.Task, error)
    CreateTask(ctx context.Context, task *models.Task) (models.Task, error)
    UpdateTask(ctx context.Context, id int, task *models.Task) (models.Task, error)
    DeleteTask(ctx context.Context, id int) error
}

type taskService struct {
    repo repositories.TaskRepository
}

func NewTaskService(repo repositories.TaskRepository) TaskService {
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
