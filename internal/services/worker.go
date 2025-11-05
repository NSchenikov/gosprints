package worker

import (
	"gosprints/internal/models"
	"fmt"
	"log"
	"strings"
)

type WorkerPool struct {
	jobs chan func()
}

func NewWorkerPool(workerCount int) *WorkerPool {
	wp := &WorkerPool{
		jobs: make(chan func(), 100),
	}

	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for job := range wp.jobs {
				job()
			}
		}(i)
	}
	return wp
}

func (wp *WorkerPool) Submit(job func()) {
	wp.jobs <- job
}


// Логика обработки. Здесь просто создаётся summary в верхнем регистре
func Processing(task models.Task) {
	summary := fmt.Sprintf("TASK #%d — %s", task.ID, strings.ToUpper(task.Text))
	log.Println("Processed:", summary)
}
