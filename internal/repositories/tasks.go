package repositories

import (
	"database/sql"
	"fmt"
	"gosprints/internal/models"
    "context"
)

type TaskRepository interface {
	GetAll(ctx context.Context) ([]models.Task, error) //нужна асинхронная обработка
	GetByID(id int) (*models.Task, error) //не нужна асинхронная обработка потому что нужно просто получить одну задачу из БД
	Create(ctx context.Context, task *models.Task) (int, error) //нужен processing, потому что появляется новая задача со статусом pending
	Update(id int, text string) error //не нужна асинхронная обработка потому что нужно просто изменить текст одной задачи
	Delete(id int) error //не нужна асинхронная обработка потому что нужно просто удалить конкретную задачу из БД
    UpdateStatus(ctx context.Context, id int, status string, startedAt, endedAt *string) error
    GetByStatus(ctx context.Context, status string) ([]models.Task, error)
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) GetAll(ctx context.Context) ([]models.Task, error) {
	rows, err := r.db.Query(`SELECT id, text, status, created_at, started_at, ended_at FROM "Tasks" ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
        var startedAt, endedAt sql.NullTime
		if err := rows.Scan(&task.ID, &task.Text, &task.Status, &task.CreatedAt, &startedAt, &endedAt); err != nil {
			return nil, err
		}

        if startedAt.Valid {
            task.StartedAt = &startedAt.Time
        }
        if endedAt.Valid {
            task.EndedAt = &endedAt.Time
        }

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *taskRepository) GetByStatus(ctx context.Context, status string) ([]models.Task, error) {
    query := `SELECT id, text, status, created_at, started_at, ended_at
              FROM "Tasks"
              WHERE status = $1
              ORDER BY id`

    rows, err := r.db.QueryContext(ctx, query, status)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []models.Task
    for rows.Next() {
    var t models.Task
    var startedAt sql.NullTime
    var endedAt   sql.NullTime

    if err := rows.Scan(
        &t.ID,
        &t.Text,
        &t.Status,
        &t.CreatedAt,
        &startedAt,
        &endedAt,
    ); err != nil {
        return nil, err
    }

    if startedAt.Valid {
        t.StartedAt = &startedAt.Time
    } else {
        t.StartedAt = nil
    }

    if endedAt.Valid {
        t.EndedAt = &endedAt.Time
    } else {
        t.EndedAt = nil
    }

    tasks = append(tasks, t)
}

    return tasks, rows.Err()
}

func (r *taskRepository) UpdateStatus(ctx context.Context, id int, status string, startedAt, endedAt *string) error {
	query := `UPDATE "Tasks" SET status=$1`
	args := []interface{}{status}
    paramCount := 1

	if startedAt != nil {
        paramCount++
        query += fmt.Sprintf(`, started_at = $%d`, paramCount)
        args = append(args, *startedAt)
	}
	if endedAt != nil {
        paramCount++
        query += fmt.Sprintf(`, ended_at = $%d`, paramCount)
        args = append(args, *endedAt)
	}
    paramCount++
    query += fmt.Sprintf(` WHERE id = $%d`, paramCount)
    args = append(args, id)
    _, err := r.db.ExecContext(ctx, query, args...)
    return err
}

func (r *taskRepository) GetByID(id int) (*models.Task, error) {
    
    var task models.Task
    err := r.db.QueryRow(`SELECT id, text FROM "Tasks" WHERE id = $1`, id).Scan(&task.ID, &task.Text)
    
    if err != nil {
        return nil, err
    }
    return &task, nil
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) (int, error) {
	    var id int
		query := `INSERT INTO "Tasks" (text, status, created_at) VALUES ($1, $2, NOW()) RETURNING id`
		err := r.db.QueryRowContext(ctx, query, task.Text, task.Status).Scan(&id)
        if err != nil {
            return 0, err
        }

		task.ID = id

		return id, nil
}

func (r *taskRepository) Update(id int, text string) error {
    
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

func (r *taskRepository) Delete(id int) error {

	result, err := r.db.Exec(`DELETE FROM "Tasks" WHERE id = $1`, id)
	if err != nil {
            fmt.Printf("Database delete error: %v\n", err)
            return err
    }

	rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return fmt.Errorf("task with ID %d not found", id)
    }
    
    fmt.Printf("Task deleted: ID=%d\n", id)

	return nil
}
