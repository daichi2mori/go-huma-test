package handler

import (
	"context"
	"encoding/json"
	"errors"
	"go-huma-test/domain"
	"go-huma-test/usecase"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type TodoHandler struct {
	uc usecase.TodoUseCase
}

func NewTodoHandler(uc usecase.TodoUseCase) *TodoHandler {
	return &TodoHandler{uc: uc}
}

type healthOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

func (h *TodoHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "health",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
	}, func(ctx context.Context, i *struct{}) (*healthOutput, error) {
		out := &healthOutput{}
		out.Body.Status = "ok"
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-todos",
		Method:      http.MethodGet,
		Path:        "/todos",
		Summary:     "List all todos",
	}, h.listTodos)

	huma.Register(api, huma.Operation{
		OperationID: "get-todo",
		Method:      http.MethodGet,
		Path:        "/todos/{id}",
		Summary:     "Get a todo by ID",
	}, h.getTodo)

	huma.Register(api, huma.Operation{
		OperationID:   "create-todo",
		Method:        http.MethodPost,
		Path:          "/todos",
		Summary:       "Create a new todo",
		DefaultStatus: http.StatusCreated,
	}, h.createTodo)

	huma.Register(api, huma.Operation{
		OperationID: "update-todo",
		Method:      http.MethodPut,
		Path:        "/todos/{id}",
		Summary:     "Update a todo",
	}, h.updateTodo)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-todo",
		Method:        http.MethodDelete,
		Path:          "/todos/{id}",
		Summary:       "Delete a todo",
		DefaultStatus: http.StatusNoContent,
	}, h.deleteTodo)
}

// --- Input / Output types ---

type listTodosOutput struct {
	Body []*domain.Todo
}

type getTodoInput struct {
	ID int64 `path:"id"`
}

type getTodoOutput struct {
	Body *domain.Todo
}

type createTodoInput struct {
	Body struct {
		Title       string  `json:"title" minLength:"1"`
		Description *string `json:"description,omitempty"`
	}
}

type createTodoOutput struct {
	Body *domain.Todo
}

type updateTodoInput struct {
	ID   int64 `path:"id"`
	Body struct {
		Title       string  `json:"title" minLength:"1"`
		Description *string `json:"description,omitempty"`
		Completed   bool    `json:"completed"`
	}
}

type updateTodoOutput struct {
	Body *domain.Todo
}

type deleteTodoInput struct {
	ID int64 `path:"id"`
}

// --- Handler methods ---

func (h *TodoHandler) getTodo(ctx context.Context, input *getTodoInput) (*getTodoOutput, error) {
	todo, err := h.uc.GetTodo(ctx, input.ID)
	if err != nil {
		if isNotFound(err) {
			return nil, huma.Error404NotFound("todo not found")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &getTodoOutput{Body: todo}, nil
}

func (h *TodoHandler) listTodos(ctx context.Context, _ *struct{}) (*listTodosOutput, error) {
	todos, err := h.uc.ListTodos(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &listTodosOutput{Body: todos}, nil
}

func (h *TodoHandler) createTodo(ctx context.Context, input *createTodoInput) (*createTodoOutput, error) {
	todo, err := h.uc.CreateTodo(ctx, domain.CreateTodoInput{
		Title:       input.Body.Title,
		Description: input.Body.Description,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &createTodoOutput{Body: todo}, nil
}

func (h *TodoHandler) updateTodo(ctx context.Context, input *updateTodoInput) (*updateTodoOutput, error) {
	todo, err := h.uc.UpdateTodo(ctx, input.ID, domain.UpdateTodoInput{
		Title:       input.Body.Title,
		Description: input.Body.Description,
		Completed:   input.Body.Completed,
	})
	if err != nil {
		if isNotFound(err) {
			return nil, huma.Error404NotFound("todo not found")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &updateTodoOutput{Body: todo}, nil
}

func (h *TodoHandler) deleteTodo(ctx context.Context, input *deleteTodoInput) (*struct{}, error) {
	if err := h.uc.DeleteTodo(ctx, input.ID); err != nil {
		if isNotFound(err) {
			return nil, huma.Error404NotFound("todo not found")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return nil, nil
}

func isNotFound(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
