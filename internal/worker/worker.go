package worker

import (
	"gosprints/internal/queue"
	"gosprints/internal/repositories"
	"fmt"
	"log"
	"math/rand"
	"time"
	"context"
)

type Worker struct {
	ID   int
	Repo repositories.TaskRepository
	Queue *queue.TaskQueue
}

func NewWorker(id int, repo repositories.TaskRepository, q *queue.TaskQueue) *Worker {
	return &Worker{ID: id, Repo: repo, Queue: q}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		log.Printf("[worker %d] started", w.ID)
        defer log.Printf("[worker %d] stopped", w.ID)

		for {
			select {
				case task, ok := <-w.Queue.Tasks():
					if !ok {
						log.Printf("[worker %d] queue closed, exit", w.ID)
						return
					}
					start := time.Now()
					log.Printf("[Worker %d] Started processing task #%d (%s) at %v\n", w.ID, task.ID, task.Text, start)

					err := w.Repo.UpdateStatus(ctx, task.ID, "processing", &start, nil)
					if err != nil {
						log.Printf("[Worker %d] Failed to update task #%d to 'processing': %v", w.ID, task.ID, err)
						continue
					}

					// имитация обработки таски
					processTime := time.Duration(rand.Intn(3)+1) * time.Second
					time.Sleep(processTime)

					end := time.Now()
					err = w.Repo.UpdateStatus(ctx, task.ID, "completed", nil, &end)
					if err != nil {
						log.Printf("[Worker %d] Failed to update task #%d to 'completed': %v", w.ID, task.ID, err)
						continue
					}

					log.Printf("[Worker %d] Completed task processing #%d (%s) at %v\n", w.ID, task.ID, task.Text, end)
					fmt.Println("----------------------------------------")
				case <-ctx.Done():
					log.Printf("[worker %d] received shutdown signal", w.ID)
                	return
			}
		}
	}()
}