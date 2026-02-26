package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go-huma-test/domain"

	dbsqlc "go-huma-test/db/sqlc"
)

type todoRepository struct {
	queries *dbsqlc.Queries
}

func NewTodoRepository(db *sql.DB) domain.TodoRepository {
	return &todoRepository{queries: dbsqlc.New(db)}
}

func (r *todoRepository) GetByID(ctx context.Context, id int64) (*domain.Todo, error) {
	row, err := r.queries.GetTodo(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("repository.GetByID: %w", err)
	}
	return toDomain(row), nil
}

func (r *todoRepository) List(ctx context.Context) ([]*domain.Todo, error) {
	rows, err := r.queries.ListTodos(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.List: %w", err)
	}
	todos := make([]*domain.Todo, len(rows))
	for i, row := range rows {
		todos[i] = toDomain(row)
	}
	return todos, nil
}

func (r *todoRepository) Create(ctx context.Context, input domain.CreateTodoInput) (*domain.Todo, error) {
	params := dbsqlc.CreateTodoParams{
		Title:       input.Title,
		Description: toNullString(input.Description),
		Completed:   0,
	}
	row, err := r.queries.CreateTodo(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("repository.Create: %w", err)
	}
	return toDomain(row), nil
}

func (r *todoRepository) Update(ctx context.Context, id int64, input domain.UpdateTodoInput) (*domain.Todo, error) {
	completed := int64(0)
	if input.Completed {
		completed = 1
	}
	params := dbsqlc.UpdateTodoParams{
		ID:          id,
		Title:       input.Title,
		Description: toNullString(input.Description),
		Completed:   completed,
	}
	row, err := r.queries.UpdateTodo(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("repository.Update: %w", err)
	}
	return toDomain(row), nil
}

func (r *todoRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeleteTodo(ctx, id); err != nil {
		return fmt.Errorf("repository.Delete: %w", err)
	}
	return nil
}

func toDomain(t dbsqlc.Todo) *domain.Todo {
	d := &domain.Todo{
		ID:        t.ID,
		Title:     t.Title,
		Completed: t.Completed != 0,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	if t.Description.Valid {
		d.Description = new(t.Description.String)
	}
	return d
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
