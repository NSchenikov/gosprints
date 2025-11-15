package repositories

import (
	"database/sql"
    "time"
	"gosprints/internal/models"
    "context"
)

type TaskRepository interface {
	GetAll(ctx context.Context) ([]models.Task, error) //нужна асинхронная обработка
	GetByID(ctx context.Context, id int) (*models.Task, error) //не нужна асинхронная обработка потому что нужно просто получить одну задачу из БД
	Create(ctx context.Context, task *models.Task) (int, error) //нужен processing, потому что появляется новая задача со статусом pending
	Update(ctx context.Context, task *models.Task) error //не нужна асинхронная обработка потому что нужно просто изменить текст одной задачи
	Delete(ctx context.Context, id int) error //не нужна асинхронная обработка потому что нужно просто удалить конкретную задачу из БД
    UpdateStatus(ctx context.Context, id int, status string, startedAt, endedAt *time.Time) error
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

func (r *taskRepository) UpdateStatus(
	ctx context.Context,
	id int,
	status string,
	startedAt, endedAt *time.Time,
) error {
	query := `UPDATE "Tasks"
	          SET status = $1,
	              started_at = $2,
	              ended_at = $3
	          WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query, status, startedAt, endedAt, id)
	return err
}

func (r *taskRepository) GetByID(ctx context.Context, id int) (*models.Task, error) {
    query := `SELECT id, text, status, created_at, started_at, ended_at
              FROM "Tasks"
              WHERE id = $1`

    row := r.db.QueryRowContext(ctx, query, id)

    var t models.Task
    var startedAt sql.NullTime
    var endedAt   sql.NullTime

    if err := row.Scan(
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
    }
    if endedAt.Valid {
        t.EndedAt = &endedAt.Time
    }

    return &t, nil
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) (int, error) {
		query := `INSERT INTO "Tasks" (text, status, created_at) VALUES ($1, $2, NOW()) RETURNING id, created_at`
        var id int
        var createdAt time.Time

		err := r.db.QueryRowContext(ctx, query, task.Text, task.Status).Scan(&id, &createdAt)
        if err != nil {
            return 0, err
        }

		task.ID = id
        task.CreatedAt = createdAt

		return id, nil
}

func (r *taskRepository) Update(ctx context.Context, task *models.Task) error {
    query := `UPDATE "Tasks"
              SET text = $1
              WHERE id = $2`

    _, err := r.db.ExecContext(ctx, query, task.Text, task.ID)
    return err
}

func (r *taskRepository) Delete(ctx context.Context, id int) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM "Tasks" WHERE id = $1`, id)
    return err
}
