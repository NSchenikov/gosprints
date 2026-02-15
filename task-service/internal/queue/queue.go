package queue

import (
	"gosprints/internal/models"
	"sync"
)

// In-memory queue
type TaskQueue struct {
	ch    chan models.Task
	store map[int]*models.Task
	mu    sync.Mutex
}

func NewTaskQueue(buffer int) *TaskQueue {
	return &TaskQueue{
		ch:    make(chan models.Task, buffer),
		store: make(map[int]*models.Task),
	}
}

func (q *TaskQueue) Add(t models.Task) {
	q.mu.Lock()
	q.store[t.ID] = &t
	q.mu.Unlock()
	q.ch <- t
}

func (q *TaskQueue) Tasks() <-chan models.Task {
	return q.ch
}

func (q *TaskQueue) GetAll() []models.Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	var result []models.Task 
	for _, t := range q.store {
		result = append(result, *t)
	}
	return result
}
