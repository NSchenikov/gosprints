package storage

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "time"
    
    _ "github.com/ClickHouse/clickhouse-go"
    "etl-worker/internal/models"
)

type AnalyticsStorage interface {
    SaveAnalytics(ctx context.Context, stats *models.TaskAnalytics) error
    GetUserStats(ctx context.Context, userID string) (*models.TaskAnalytics, error)
    Close() error
}

type ClickHouseStorage struct {
    db *sql.DB
}

func NewClickHouseStorage(host string, port int, database string) (*ClickHouseStorage, error) {
    connStr := fmt.Sprintf("tcp://%s:%d/%s?username=default&password=&compress=true", 
        host, port, database)
    
    db, err := sql.Open("clickhouse", connStr)
    if err != nil {
        return nil, err
    }
    
    // Создаём таблицу, если её нет
    err = p.createTable()
    if err != nil {
        return nil, err
    }
    
    return &ClickHouseStorage{db: db}, nil
}

func (s *ClickHouseStorage) createTable() error {
    query := `
    CREATE TABLE IF NOT EXISTS task_analytics (
        user_id String,
        tasks_created Int32,
        tasks_completed Int32,
        avg_completion_time Float64,
        last_event_time DateTime,
        date Date
    ) ENGINE = MergeTree()
    ORDER BY (user_id, date)
    `
    _, err := s.db.Exec(query)
    return err
}

func (s *ClickHouseStorage) SaveAnalytics(ctx context.Context, stats *models.TaskAnalytics) error {
    query := `
    INSERT INTO task_analytics (user_id, tasks_created, tasks_completed, avg_completion_time, last_event_time, date)
    VALUES (?, ?, ?, ?, ?, ?)
    `
    
    _, err := s.db.ExecContext(ctx, query,
        stats.UserId,
        stats.TasksCreated,
        stats.TasksCompleted,
        stats.AvgCompletionTime,
        stats.LastEventTime,
        stats.Date,
    )
    
    if err != nil {
        log.Printf("[ClickHouse] Insert error: %v", err)
        return err
    }
    
    log.Printf("[ClickHouse] Saved analytics for user %s", stats.UserId)
    return nil
}

func (s *ClickHouseStorage) GetUserStats(ctx context.Context, userID string) (*models.TaskAnalytics, error) {
    query := `
    SELECT user_id, tasks_created, tasks_completed, avg_completion_time, last_event_time, date
    FROM task_analytics
    WHERE user_id = ? AND date = today()
    ORDER BY last_event_time DESC
    LIMIT 1
    `
    
    var stats models.TaskAnalytics
    err := s.db.QueryRowContext(ctx, query, userID).Scan(
        &stats.UserId,
        &stats.TasksCreated,
        &stats.TasksCompleted,
        &stats.AvgCompletionTime,
        &stats.LastEventTime,
        &stats.Date,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    
    return &stats, nil
}

func (s *ClickHouseStorage) Close() error {
    return s.db.Close()
}