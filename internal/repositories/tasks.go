package repositories

import (
	"database/sql"
	"fmt"
	"gosprints/internal/models"
)

type TaskRepository interface {
	GetAll() ([]models.Task, error)
	GetByID(id int) (*models.Task, error)
	Create(task *models.Task) error
	Update(id int, text string) error
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

func (r *taskRepository) Create(task *models.Task) error {
	    err := r.db.QueryRow(`INSERT INTO "Tasks" (text) VALUES ($1) RETURNING id`, task.Text).Scan(&task.ID)
        if err != nil {
            return err
        }

		return nil
}

func (r *taskRepository) Update(id int, text string) error {
    fmt.Printf("🎯 REPOSITORY Update: id=%d, text=%s\n", id, text)
    
    result, err := r.db.Exec(
        `UPDATE "Tasks" SET text = $1 WHERE id = $2`,
        text, id,
    )
    if err != nil {
        fmt.Printf("Update ERROR: %v\n", err)
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return fmt.Errorf("task with ID %d not found", id)
    }
    
    fmt.Printf("Task updated: ID=%d\n", id)
    return nil
}