package metrics

import (
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex
	
	// Задачи
	tasksCreated   int64
	tasksCompleted int64
	tasksFailed    int64
	
	// Время обработки
	processingTimeTotal time.Duration
	processingCount     int64
	
	// WebSocket
	wsConnectionsActive int64
	wsConnectionsTotal  int64
	
	// API запросы
	apiRequestsTotal int64
	apiErrorsTotal   int64
	
	// Время запуска для uptime
	startTime time.Time
}

var globalMetrics = &Metrics{
	startTime: time.Now(),
}

// Геттеры
func Get() *Metrics {
	return globalMetrics
}

// Задачи
func (m *Metrics) IncTasksCreated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasksCreated++
}

func (m *Metrics) IncTasksCompleted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasksCompleted++
}

func (m *Metrics) IncTasksFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasksFailed++
}

// Время обработки
func (m *Metrics) AddProcessingTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processingTimeTotal += duration
	m.processingCount++
}

// WebSocket
func (m *Metrics) IncWSConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.wsConnectionsActive++
	m.wsConnectionsTotal++
}

func (m *Metrics) DecWSConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.wsConnectionsActive > 0 {
		m.wsConnectionsActive--
	}
}

func (m *Metrics) GetWSConnectionsActive() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.wsConnectionsActive
}

// API
func (m *Metrics) IncAPIRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.apiRequestsTotal++
}

func (m *Metrics) IncAPIErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.apiErrorsTotal++
}

// Рассчетные метрики
func (m *Metrics) AvgProcessingTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.processingCount == 0 {
		return 0
	}
	return m.processingTimeTotal / time.Duration(m.processingCount)
}

func (m *Metrics) Uptime() time.Duration {
	return time.Since(m.startTime)
}

func (m *Metrics) TasksTotal() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tasksCreated
}

func (m *Metrics) APIErrorRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.apiRequestsTotal == 0 {
		return 0
	}
	return float64(m.apiErrorsTotal) / float64(m.apiRequestsTotal) * 100
}