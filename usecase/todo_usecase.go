package usecase

import (
	"context"
	"fmt"
	"go-huma-test/domain"
)

type TodoUseCase interface {
	GetTodo(ctx context.Context, id int64) (*domain.Todo, error)
	ListTodos(ctx context.Context) ([]*domain.Todo, error)
	CreateTodo(ctx context.Context, input domain.CreateTodoInput) (*domain.Todo, error)
	UpdateTodo(ctx context.Context, id int64, input domain.UpdateTodoInput) (*domain.Todo, error)
	DeleteTodo(ctx context.Context, id int64) error
}

type todoUseCase struct {
	repo domain.TodoRepository
}

func NewTodoUseCase(repo domain.TodoRepository) TodoUseCase {
	return &todoUseCase{repo: repo}
}

func (u *todoUseCase) GetTodo(ctx context.Context, id int64) (*domain.Todo, error) {
	todo, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetTodo: %w", err)
	}
	return todo, nil
}

func (u *todoUseCase) ListTodos(ctx context.Context) ([]*domain.Todo, error) {
	todos, err := u.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase.ListTodos: %w", err)
	}
	return todos, nil
}

func (u *todoUseCase) CreateTodo(ctx context.Context, input domain.CreateTodoInput) (*domain.Todo, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("usecase.Create: title is required")
	}
	todo, err := u.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("usecase.CreateTodo: %w", err)
	}
	return todo, nil
}

func (u *todoUseCase) UpdateTodo(ctx context.Context, id int64, input domain.UpdateTodoInput) (*domain.Todo, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("usecase.UpdateTodo: title is required")
	}
	todo, err := u.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("usecase.UpdateTodo: %w", err)
	}
	return todo, nil
}

func (u *todoUseCase) DeleteTodo(ctx context.Context, id int64) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase.DeleteTodo: %w", err)
	}
	return nil
}
