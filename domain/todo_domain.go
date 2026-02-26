package domain

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type Todo struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTodoInput struct {
	Title       string
	Description *string
}

type UpdateTodoInput struct {
	Title       string
	Description *string
	Completed   bool
}

type TodoRepository interface {
	GetByID(ctx context.Context, id int64) (*Todo, error)
	List(ctx context.Context) ([]*Todo, error)
	Create(ctx context.Context, input CreateTodoInput) (*Todo, error)
	Update(ctx context.Context, id int64, input UpdateTodoInput) (*Todo, error)
	Delete(ctx context.Context, id int64) error
}
