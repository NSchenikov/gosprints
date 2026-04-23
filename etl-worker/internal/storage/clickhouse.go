
package storage

import (
    "context"
    "log"
    
    "github.com/ClickHouse/clickhouse-go/v2"
    "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
    
    "etl-worker/internal/models"
)

type AnalyticsStorage interface {
    SaveAnalytics(ctx context.Context, stats *models.TaskAnalytics) error
    Close() error
}

type ClickHouseStorage struct {
    conn driver.Conn
}

func NewClickHouseStorage(addr string, database string) (*ClickHouseStorage, error) {
    conn, err := clickhouse.Open(&clickhouse.Options{
        Addr: []string{addr},
        Auth: clickhouse.Auth{
            Database: database,
            Username: "default",
            Password: "clickhouse",
        },
    })
    if err != nil {
        return nil, err
    }
    
    // Создаём таблицу для агрегированных данных c ReplacingMergeTree
    err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS task_analytics (
            user_id String,
            tasks_completed Int32,
            avg_completion_time Float64,
            last_event_time DateTime,
            date Date
        ) ENGINE = ReplacingMergeTree(last_event_time)
        ORDER BY (user_id)
    `)
    if err != nil {
        return nil, err
    }
    
    return &ClickHouseStorage{conn: conn}, nil
}

func (s *ClickHouseStorage) SaveAnalytics(ctx context.Context, stats *models.TaskAnalytics) error {
    err := s.conn.Exec(ctx, `
        INSERT INTO task_analytics (user_id, tasks_completed, avg_completion_time, last_event_time, date)
        VALUES (?, ?, ?, ?, ?)
    `,
        stats.UserId,
        stats.TasksCompleted,
        stats.AvgCompletionTime,
        stats.LastEventTime,
        stats.Date,
    )
    
    if err != nil {
        log.Printf("[ClickHouse] Insert error: %v", err)
        return err
    }
    
    log.Printf("[ClickHouse] Saved analytics for user %s: completed=%d, avg=%.2f", 
        stats.UserId, stats.TasksCompleted, stats.AvgCompletionTime)
    return nil
}

func (s *ClickHouseStorage) Close() error {
    return s.conn.Close()
}