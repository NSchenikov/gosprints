package models

import (
	"time"
)

type Task struct {
    ID   int    `json:"id"`
    Text string `json:"text"`
    Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
    StartedAt *time.Time  `json:"started_at,omitempty"`
    EndedAt   *time.Time  `json:"ended_at,omitempty"`
	UserID    string     `json:"user_id"`
	UpdatedAt  time.Time  `json:"updated_at"`
	
	//для state machine
	Attempts      int        `json:"attempts"`
    Validation1At *time.Time `json:"validation1_at,omitempty"`
    Validation2At *time.Time `json:"validation2_at,omitempty"`
    ClosedAt      *time.Time `json:"closed_at,omitempty"`
}

type TaskStatusEvent struct {
	Type      string    `json:"type"`
	TaskID    int       `json:"task_id"`
	Text     string    `json:"text"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

//новые статусы
const (
	TaskStatusPending				= "PENDING"
	TaskStatusProcessing			= "PROCESSING"
	TaskStatusCompleted				= "COMPLETED"
	TaskStatusFailed				= "FAILED"

    TaskStatusNew                   = "NEW"
    TaskStatusValidation1           = "VALIDATION_1"
    TaskStatusWaitingForValidation2 = "WAITING_FOR_VALIDATION_2"
    TaskStatusValidation2           = "VALIDATION_2"
    TaskStatusReadyForClosure       = "READY_FOR_CLOSURE"
    TaskStatusClosed                = "CLOSED"
)