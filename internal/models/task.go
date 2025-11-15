
package models

import (
	"time"
)

type Task struct {
    ID   int    `json:"id"`
    Text string `json:"text"`
    Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
    StartedAt *time.Time `json:"started_at"`
    EndedAt   *time.Time `json:"ended_at"`
}