package repositories

import (
	"context"
	"gosprints/internal/models"
	"time"
)

type TaskRepository interface {
	GetAll(ctx context.Context) ([]models.Task, error)
	GetByStatus(ctx context.Context, status string) ([]models.Task, error)
	GetByID(ctx context.Context, id int) (*models.Task, error)
	Create(ctx context.Context, task *models.Task) (int, error)
	Update(ctx context.Context, task *models.Task) error
	UpdateStatus(ctx context.Context, id int, status string, startedAt, endedAt *time.Time) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter TaskFilter) ([]models.Task, int, error)
}