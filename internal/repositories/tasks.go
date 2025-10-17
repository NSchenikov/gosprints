package repositories

import (
	"database/sql"
	"gosprints/internal/models"
)

type TaskRepository interface {
	GetAll() ([]models.Task, error)
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) GetAll() ([]models.Task, error) {
	rows, err := r.db.Query("SELECT id, text FROM \"Tasks\"")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Text); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}