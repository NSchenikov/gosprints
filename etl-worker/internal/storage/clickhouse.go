
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

func NewClickHouseStorage(host string, port int, database string) (*ClickHouseStorage, error) {
    conn, err := clickhouse.Open(&clickhouse.Options{
        Addr: []string{host},
        Auth: clickhouse.Auth{
            Database: database,
            Username: "default",
            Password: "",
        },
    })
    if err != nil {
        return nil, err
    }
    
    // Создаём таблицу
    err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS task_analytics (
            user_id String,
            tasks_created Int32,
            tasks_completed Int32,
            avg_completion_time Float64,
            last_event_time DateTime,
            date Date
        ) ENGINE = MergeTree()
        ORDER BY (user_id, date)
    `)
    if err != nil {
        return nil, err
    }
    
    return &ClickHouseStorage{conn: conn}, nil
}

func (s *ClickHouseStorage) SaveAnalytics(ctx context.Context, stats *models.TaskAnalytics) error {
    err := s.conn.Exec(ctx, `
        INSERT INTO task_analytics (user_id, tasks_created, tasks_completed, avg_completion_time, last_event_time, date)
        VALUES (?, ?, ?, ?, ?, ?)
    `,
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

func (s *ClickHouseStorage) Close() error {
    return s.conn.Close()
}