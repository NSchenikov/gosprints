package scheduler

import (
    "context"
    "log"
    "time"

    "gosprints/internal/models"
)

type TaskRepository interface {
	GetByStatus(ctx context.Context, status string) ([]models.Task, error)
}

type TaskQueue interface {
	Add(task models.Task)
}

type Dispatcher struct {
    repo     TaskRepository
    queue    TaskQueue
    interval time.Duration
}

func NewDispatcher(repo TaskRepository, q TaskQueue, interval time.Duration) *Dispatcher {
    return &Dispatcher{
        repo:     repo,
        queue:    q,
        interval: interval,
    }
}

func (d *Dispatcher) Start(ctx context.Context) {
    go func() {

        log.Println("[dispatcher] started")
        defer log.Println("[dispatcher] stopped")

        ticker := time.NewTicker(d.interval)
        defer ticker.Stop()

        for {
            select {
                case <-ticker.C:
                    dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
                    tasks, err := d.repo.GetByStatus(dbCtx, "pending")
                    cancel()
                    if err != nil {
                        log.Printf("[dispatcher] failed to get pending tasks: %v\n", err)
                        continue
                    }

                    if len(tasks) == 0 {
                        continue
                    }

                    log.Printf("[dispatcher] found %d pending tasks\n", len(tasks))

                    for _, t := range tasks {
                        d.queue.Add(t)
                    }
                case <-ctx.Done():
                    log.Println("[dispatcher] received shutdown signal")
                    return
            }
        }
    }()
}
