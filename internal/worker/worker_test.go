package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"gosprints/internal/models"
)

type statusCall struct {
	ID        int
	Status    string
	StartedAt *time.Time
	EndedAt   *time.Time
}

// фейковая версия TaskRepository
type fakeRepo struct {
	mu    sync.Mutex
	calls []statusCall
}

func (f *fakeRepo) UpdateStatus(ctx context.Context, id int, status string, startedAt, endedAt *time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, statusCall{
		ID:        id,
		Status:    status,
		StartedAt: startedAt,
		EndedAt:   endedAt,
	})
	return nil
}

func (f *fakeRepo) Calls() []statusCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]statusCall, len(f.calls))
	copy(out, f.calls)
	return out
}

func (f *fakeRepo) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

// фейковая версия TaskQueue
type fakeQueue struct {
	ch chan models.Task
}

func newFakeQueue(buf int) *fakeQueue {
	return &fakeQueue{
		ch: make(chan models.Task, buf),
	}
}

//
func (q *fakeQueue) Tasks() <-chan models.Task {
	return q.ch
}

// хелпер, позволяющий запушить задачу в очередь
func (q *fakeQueue) push(task models.Task) {
	q.ch <- task
}

// ждём, пока переданная функция либо вернёт true, либо пока закончится время
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}

func TestWorker_ChangesTaskStatusProcessingToCompleted(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo := &fakeRepo{}
	q := newFakeQueue(1)

	// воркер с фейковыми зависимостями
	w := NewWorker(1, repo, q)
	w.Start(ctx)

	// имитируем работу диспетчера
	task := models.Task{
		ID:     42,
		Text:   "test worker",
		Status: "pending",
	}
	q.push(task)

	// ждём, пока воркер хотя бы пару раз вызовет UpdateStatus
	waitFor(t, 5*time.Second, func() bool {
		return repo.CallCount() >= 2
	})

	calls := repo.Calls()
	if len(calls) < 2 {
		t.Fatalf("expected at least 2 UpdateStatus calls, got %d", len(calls))
	}

	// проверяем, что оба вызова относятся к той же задаче
	if calls[0].ID != task.ID || calls[1].ID != task.ID {
		t.Fatalf("expected calls for task %d, got IDs: %d, %d",
			task.ID, calls[0].ID, calls[1].ID)
	}

	// проверяем первый статус — processing?
	if calls[0].Status != "processing" {
		t.Errorf("expected first status 'processing', got '%s'", calls[0].Status)
	}

	// проверяем второй статус — completed?
	if calls[1].Status != "completed" {
		t.Errorf("expected second status 'completed', got '%s'", calls[1].Status)
	}

	if calls[0].StartedAt == nil {
		t.Errorf("expected StartedAt to be set for 'processing' status")
	}
	if calls[1].EndedAt == nil {
		t.Errorf("expected EndedAt to be set for 'completed' status")
	}
}
