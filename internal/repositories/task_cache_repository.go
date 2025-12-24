package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gosprints/internal/cache"
	"gosprints/internal/models"
	"gosprints/internal/services"
)

type TaskCacheRepository struct {
	baseRepo *taskRepository
	cache    cache.Cache
	stats    cache.CacheStats
	mu       sync.RWMutex
}

const (
	CacheTTLAllTasks      = 1 * time.Minute    // Короткий TTL для списков
	CacheTTLSingleTask    = 5 * time.Minute    // Средний TTL для одной задачи
	CacheTTLByStatus      = 2 * time.Minute    // TTL для задач по статусу
)

func NewTaskCacheRepository(baseRepo *taskRepository, cache cache.Cache) *TaskCacheRepository {
	return &TaskCacheRepository{
		baseRepo: baseRepo,
		cache:    cache,
		// stats:    cache.CacheStats{},
	}
}

//key generators
func (r *TaskCacheRepository) allTasksKey() string {
	return "tasks:all"
}

func (r *TaskCacheRepository) taskByIDKey(id int) string {
	return fmt.Sprintf("task:%d", id)
}

func (r *TaskCacheRepository) tasksByStatusKey(status string) string {
	return fmt.Sprintf("tasks:status:%s", status)
}

//отчасти повторяем все методы, но с кэшированием
func (r *TaskCacheRepository) GetAll(ctx context.Context) ([]models.Task, error) {
	cacheKey := r.allTasksKey()
	
	// Пытаемся получить из кэша
	if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		r.incHit()
		var tasks []models.Task
		if err := json.Unmarshal(cached, &tasks); err == nil {
			return tasks, nil
		}
	}
	
	r.incMiss()
	
	// берем из базы
	tasks, err := r.baseRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	
	// кладем в кэш
	if tasks != nil {
		if data, err := json.Marshal(tasks); err == nil {
			r.cache.Set(ctx, cacheKey, data, CacheTTLAllTasks)
			r.mu.Lock()
			r.stats.Sets++
			r.mu.Unlock()
		}
	}
	
	return tasks, nil
}

func (r *TaskCacheRepository) GetByID(ctx context.Context, id int) (*models.Task, error) {
	cacheKey := r.taskByIDKey(id)
	
	// Пытаемся достать из кэша
	if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		r.incHit()
		var task models.Task
		if err := json.Unmarshal(cached, &task); err == nil {
			return &task, nil
		}
	}
	
	r.incMiss()
	
	// берем из базы
	task, err := r.baseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Сохраняем в кэше
	if task != nil {
		if data, err := json.Marshal(task); err == nil {
			r.cache.Set(ctx, cacheKey, data, CacheTTLSingleTask)
			r.mu.Lock()
			r.stats.Sets++
			r.mu.Unlock()
		}
	}
	
	return task, nil
}

func (r *TaskCacheRepository) GetByStatus(ctx context.Context, status string) ([]models.Task, error) {
	cacheKey := r.tasksByStatusKey(status)
	
	// Пытаемся получить из кэша
	if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		r.incHit()
		var tasks []models.Task
		if err := json.Unmarshal(cached, &tasks); err == nil {
			return tasks, nil
		}
	}
	
	r.incMiss()
	
	// Получаем из базы
	tasks, err := r.baseRepo.GetByStatus(ctx, status)
	if err != nil {
		return nil, err
	}
	
	// Сохраняем в кэш
	if tasks != nil {
		if data, err := json.Marshal(tasks); err == nil {
			r.cache.Set(ctx, cacheKey, data, CacheTTLByStatus)
			r.mu.Lock()
			r.stats.Sets++
			r.mu.Unlock()
		}
	}
	
	return tasks, nil
}

// Create с инвалидацией кэша
func (r *TaskCacheRepository) Create(ctx context.Context, task *models.Task) (int, error) {
	id, err := r.baseRepo.Create(ctx, task)
	if err != nil {
		return 0, err
	}
	
	go r.invalidateCache(ctx)
	
	return id, nil
}

// Update с инвалидацией кэша
func (r *TaskCacheRepository) Update(ctx context.Context, task *models.Task) error {

	oldTask, err := r.baseRepo.GetByID(ctx, task.ID)
	if err != nil {
		return err
	}
	
	// Обновляем
	err = r.baseRepo.Update(ctx, task)
	if err != nil {
		return err
	}
	
	// Инвалидируем кэш асинхронно
	go r.invalidateTaskCache(context.Background(), task.ID, oldTask)
	
	return nil
}

// UpdateStatus с инвалидацией кэша
func (r *TaskCacheRepository) UpdateStatus(
	ctx context.Context,
	id int,
	status string,
	startedAt, endedAt *time.Time,
) error {

	oldTask, err := r.baseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	// Обновляем
	err = r.baseRepo.UpdateStatus(ctx, id, status, startedAt, endedAt)
	if err != nil {
		return err
	}
	
	// Инвалидируем
	go func() {
		ctx := context.Background()
		r.invalidateTaskCache(ctx, id, oldTask)
	}()
	
	return nil
}

// Delete с инвалидацией кэша
func (r *TaskCacheRepository) Delete(ctx context.Context, id int) error {
	
	task, err := r.baseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	// Удаляем
	err = r.baseRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	
	// Инвалидация
	go func() {
		ctx := context.Background()
		r.invalidateTaskCache(ctx, id, task)
	}()
	
	return nil
}

// Вспомогательные методы для инвалидации
func (r *TaskCacheRepository) invalidateCache(ctx context.Context) {

	r.cache.Delete(ctx, r.allTasksKey())
	r.cache.InvalidateByPattern(ctx, "tasks:status:")
}

func (r *TaskCacheRepository) invalidateTaskCache(ctx context.Context, id int, oldTask *models.Task) {
	
	r.cache.Delete(ctx, r.taskByIDKey(id))
	
	
	r.cache.Delete(ctx, r.allTasksKey())
	
	
	if oldTask != nil {
		r.cache.Delete(ctx, r.tasksByStatusKey(oldTask.Status))
	}
}

// Методы для управления кэшем
func (r *TaskCacheRepository) WarmUpCache(ctx context.Context) error {
	// Загружаем все задачи
	tasks, err := r.baseRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	
	// Кэшируем все задачи по отдельности
	for _, task := range tasks {
		cacheKey := r.taskByIDKey(task.ID)
		if data, err := json.Marshal(task); err == nil {
			r.cache.Set(ctx, cacheKey, data, CacheTTLSingleTask)
		}
	}
	
	// Кэшируем по статусам
	statuses := []string{"pending", "in_progress", "completed"}
	for _, status := range statuses {
		tasksByStatus, err := r.baseRepo.GetByStatus(ctx, status)
		if err != nil {
			continue
		}
		
		cacheKey := r.tasksByStatusKey(status)
		if data, err := json.Marshal(tasksByStatus); err == nil {
			r.cache.Set(ctx, cacheKey, data, CacheTTLByStatus)
		}
	}
	
	return nil
}

func (r *TaskCacheRepository) ClearCache(ctx context.Context) error {
	r.cache.InvalidateByPattern(ctx, "task:")
	r.cache.InvalidateByPattern(ctx, "tasks:")
	
	r.mu.Lock()
    r.stats = cache.CacheStats{}
    r.mu.Unlock()

	return r.cache.InvalidateByPattern(ctx, "task:")
}

func (r *TaskCacheRepository) GetCacheStats() cache.CacheStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stats
}

func (r *TaskCacheRepository) incHit() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Hits++
}

func (r *TaskCacheRepository) incMiss() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Misses++
}

var _ services.TaskRepository = (*TaskCacheRepository)(nil)