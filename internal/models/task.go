
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
}

type TaskStatusEvent struct {
	Type      string    `json:"type"`
	TaskID    int       `json:"book_id"`
	Text     string    `json:"text"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}