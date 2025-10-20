package repositories

import (
	"database/sql"
	// "fmt"
	"gosprints/internal/models"
)

type TaskRepository interface {
	GetAll() ([]models.Task, error)
	GetByID(id int) (*models.Task, error)
	// Create(task *models.Task) error
	// Update(id int, text string) error
	// Delete(id int) error
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) GetAll() ([]models.Task, error) {
	rows, err := r.db.Query(`SELECT id, text FROM "Tasks"`)
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

func (r *taskRepository) GetByID(id int) (*models.Task, error) {
    
    var task models.Task
    err := r.db.QueryRow(`SELECT id, text FROM "Tasks" WHERE id = $1`, id).Scan(&task.ID, &task.Text)
    
    if err != nil {
        return nil, err
    }
    return &task, nil
}