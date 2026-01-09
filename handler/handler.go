package handler

import (
	"context"
	"database/sql"
	"fmt"
	"go-huma-test/db"
	"go-huma-test/model"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type TodoHandler struct {
	queries *db.Queries
	db      *sql.DB
}

func NewTodoHandler(queries *db.Queries, db *sql.DB) *TodoHandler {
	return &TodoHandler{
		queries: queries,
		db:      db,
	}
}

func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func nullStringToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func toTodoResponse(t db.Todo) model.TodoResponse {
	description := nullStringToString(t.Description)

	return model.TodoResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: &description,
		Completed:   t.Completed == 1,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *TodoHandler) ListTodos(ctx context.Context, input *model.ListTodosInput) (*model.ListTodosOutput, error) {
	var todos []db.Todo
	var err error

	if input.Completed {
		todos, err = h.queries.ListTodosByStatus(ctx, 1)
	} else {
		todos, err = h.queries.ListTodos(ctx)
	}

	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch todos", err)
	}

	output := &model.ListTodosOutput{}
	output.Body.Todos = make([]model.TodoResponse, len(todos))
	for i, t := range todos {
		output.Body.Todos[i] = toTodoResponse(t)
	}

	return output, nil
}

func (h *TodoHandler) GetTodo(ctx context.Context, input *model.GetTodoInput) (*model.GetTodoOutput, error) {
	todo, err := h.queries.GetTodo(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma.Error404NotFound(fmt.Sprintf("Todo with ID %d not found", input.ID))
		}
		return nil, huma.Error500InternalServerError("Failed to fetch todo", err)
	}

	return &model.GetTodoOutput{Body: toTodoResponse(todo)}, nil
}

func (h *TodoHandler) CreateTodo(ctx context.Context, input *model.CreateTodoInput) (*model.CreateTodoOutput, error) {
	description := stringToNullString(*input.Body.Description)

	todo, err := h.queries.CreateTodo(ctx, db.CreateTodoParams{
		Title:       input.Body.Title,
		Description: description,
		Completed:   0,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create todo", err)
	}

	return &model.CreateTodoOutput{Body: toTodoResponse(todo)}, nil
}

func (h *TodoHandler) UpdateTodo(ctx context.Context, input *model.UpdateTodoInput) (*model.UpdateTodoOutput, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to begin transaction", err)
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	var completed int64
	if input.Body.Completed {
		completed = 1
	}

	description := stringToNullString(*input.Body.Description)

	todo, err := qtx.UpdateTodo(ctx, db.UpdateTodoParams{
		ID:          input.ID,
		Title:       input.Body.Title,
		Description: description,
		Completed:   completed,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update todo", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, huma.Error500InternalServerError("Failed to commit transaction", err)
	}

	return &model.UpdateTodoOutput{Body: toTodoResponse(todo)}, nil
}

func (h *TodoHandler) DeleteTodo(ctx context.Context, input *model.DeleteTodoInput) (*model.DeleteTodoOutput, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to begin transaction", err)
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	if err := qtx.DeleteTodo(ctx, input.ID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete todo", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, huma.Error500InternalServerError("Failed to commit transaction", err)
	}

	output := &model.DeleteTodoOutput{}
	output.Body.Message = "Todo deleted successfully"
	return output, nil
}

func (h *TodoHandler) ToggleTodo(ctx context.Context, input *model.ToggleTodoInput) (*model.ToggleTodoOutput, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to begin transaction", err)
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	todo, err := qtx.ToggleTodoCompleted(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to toggle todo", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, huma.Error500InternalServerError("Failed to commit transaction", err)
	}

	return &model.ToggleTodoOutput{Body: toTodoResponse(todo)}, nil
}
