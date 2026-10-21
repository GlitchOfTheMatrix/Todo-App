package models

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	Id        uuid.UUID `db:"id" json:"id"`
	UserId    uuid.UUID `db:"user_id" json:"user_id"`
	Task      string    `db:"task" json:"task"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Request DTOs

type CreateTaskRequest struct {
	Task string `json:"task"`
}

type UpdateTaskRequest struct {
	Task   *string `json:"task,omitempty"`
	Status *string `json:"status,omitempty"`
}
