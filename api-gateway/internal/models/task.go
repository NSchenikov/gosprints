package models

import "time"

type Task struct {
    ID        int       `json:"id"`
    Text      string    `json:"text"`
    Status    string    `json:"status"`
    UserID    string    `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
}
