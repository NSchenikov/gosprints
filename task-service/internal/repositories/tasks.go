package repositories

import (
	"database/sql"
    "time"
	"gosprints/internal/models"
    "context"
    "strconv"
    "fmt"
)

type taskRepository struct {
	db *sql.DB
}

type TaskFilter struct {
	UserID string
	Status string
	Page   int
	Limit  int
}

func NewTaskRepository(db *sql.DB) *taskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) GetAll(ctx context.Context) ([]models.Task, error) {
	rows, err := r.db.Query(`SELECT id, text, status, created_at, started_at, ended_at, user_id FROM "Tasks" ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
        var startedAt, endedAt sql.NullTime
		if err := rows.Scan(&task.ID, &task.Text, &task.Status, &task.CreatedAt, &startedAt, &endedAt, &task.UserID); err != nil {
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
    query := `SELECT id, text, status, created_at, started_at, ended_at, user_id
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
        &t.UserID,
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
    query := `SELECT id, text, status, created_at, started_at, ended_at, user_id
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
        &t.UserID,
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
		query := `INSERT INTO "Tasks" (text, status, user_id, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id, created_at`
        var id int
        var createdAt time.Time

		err := r.db.QueryRowContext(ctx, query, task.Text, task.Status, task.UserID).Scan(&id, &createdAt)
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

func (r *taskRepository) List(ctx context.Context, filter TaskFilter) ([]models.Task, int, error) {
	// Базовый запрос
	query := `SELECT id, text, status, created_at, started_at, ended_at, user_id 
	          FROM "Tasks" WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM "Tasks" WHERE 1=1`
	
	var args []interface{}
	var countArgs []interface{}
	argIndex := 1
	
	// фильтры
	if filter.UserID != "" {
		query += " AND user_id = $" + string('0'+byte(argIndex))
		countQuery += " AND user_id = $" + string('0'+byte(argIndex))
		args = append(args, filter.UserID)
		countArgs = append(countArgs, filter.UserID)
		argIndex++
	}
	
	if filter.Status != "" {
		query += " AND status = $" + string('0'+byte(argIndex))
		countQuery += " AND status = $" + string('0'+byte(argIndex))
		args = append(args, filter.Status)
		countArgs = append(countArgs, filter.Status)
		argIndex++
	}
	
	// сортировка и пагинация
	query += " ORDER BY created_at DESC"
	
	if filter.Limit > 0 {
		query += " LIMIT $" + string('0'+byte(argIndex))
		args = append(args, filter.Limit)
		argIndex++
		
		if filter.Page > 1 {
			offset := (filter.Page - 1) * filter.Limit
			query += " OFFSET $" + string('0'+byte(argIndex))
			args = append(args, offset)
			argIndex++
		}
	}
	
	// запрос для получения задач
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var startedAt sql.NullTime
		var endedAt sql.NullTime
		
		if err := rows.Scan(
			&t.ID,
			&t.Text,
			&t.Status,
			&t.CreatedAt,
			&startedAt,
			&endedAt,
			&t.UserID,
		); err != nil {
			return nil, 0, err
		}
		
		if startedAt.Valid {
			t.StartedAt = &startedAt.Time
		}
		if endedAt.Valid {
			t.EndedAt = &endedAt.Time
		}
		
		tasks = append(tasks, t)
	}
	
	// Получаем общее количество
	var total int
	err = r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	return tasks, total, nil
}

func (r *taskRepository) Search(ctx context.Context, query, userID string, page, limit int) ([]models.Task, int, error) {
    // Базовый запрос с полнотекстовым поиском
    sqlQuery := `
        SELECT id, text, status, created_at, started_at, ended_at, user_id,
               ts_rank(search_vector, plainto_tsquery('russian', $1)) as rank
        FROM "Tasks"
        WHERE search_vector @@ plainto_tsquery('russian', $1)
    `
    countQuery := `
        SELECT COUNT(*)
        FROM "Tasks"
        WHERE search_vector @@ plainto_tsquery('russian', $1)
    `
    
    args := []interface{}{query}
    countArgs := []interface{}{query}
    
    // фильтр по user_id если указан
    if userID != "" {
        sqlQuery += " AND user_id = $" + strconv.Itoa(len(args)+1)
        countQuery += " AND user_id = $" + strconv.Itoa(len(countArgs)+1)
        args = append(args, userID)
        countArgs = append(countArgs, userID)
    }
    
    // Сортировка по релевантности
    sqlQuery += " ORDER BY rank DESC"
    
    // Пагинация
    if limit > 0 {
        sqlQuery += " LIMIT $" + strconv.Itoa(len(args)+1)
        args = append(args, limit)
        
        if page > 1 {
            offset := (page - 1) * limit
            sqlQuery += " OFFSET $" + strconv.Itoa(len(args)+1)
            args = append(args, offset)
        }
    }
    
    // поиск
    rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("search query error: %w", err)
    }
    defer rows.Close()
    
    var tasks []models.Task
    for rows.Next() {
        var t models.Task
        var startedAt, endedAt sql.NullTime
        var rank float64
        
        if err := rows.Scan(
            &t.ID, &t.Text, &t.Status, &t.CreatedAt,
            &startedAt, &endedAt, &t.UserID, &rank,
        ); err != nil {
            return nil, 0, fmt.Errorf("scan error: %w", err)
        }
        
        if startedAt.Valid {
            t.StartedAt = &startedAt.Time
        }
        if endedAt.Valid {
            t.EndedAt = &endedAt.Time
        }
        
        tasks = append(tasks, t)
    }
    
    // общее количество
    var total int
    err = r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("count query error: %w", err)
    }
    
    return tasks, total, nil
}
