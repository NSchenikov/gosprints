package scheduler

import (
    "context"
    "log"
    "time"

    "gosprints/internal/queue"
    "gosprints/internal/repositories"
)

type Dispatcher struct {
    repo     repositories.TaskRepository
    queue    *queue.TaskQueue
    interval time.Duration
}

func NewDispatcher(repo repositories.TaskRepository, q *queue.TaskQueue, interval time.Duration) *Dispatcher {
    return &Dispatcher{
        repo:     repo,
        queue:    q,
        interval: interval,
    }
}

func (d *Dispatcher) Start() {
    go func() {
        ticker := time.NewTicker(d.interval)
        defer ticker.Stop()

        for range ticker.C {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            tasks, err := d.repo.GetByStatus(ctx, "pending")
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
        }
    }()
}
