// Package handler はTodo管理APIのHTTPリクエストハンドラーを提供する。
// このパッケージはHumaフレームワークと連携し、
// Todoの作成、取得、更新、削除などの操作を処理する。
package handler

import (
	"context"
	"database/sql"
	"fmt"
	dbsqlc "go-huma-test/db/sqlc"
	"go-huma-test/model"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// TodoHandler はTodoに関する操作を処理するハンドラー
type TodoHandler struct {
	queries *dbsqlc.Queries
	db      *sql.DB
}

// NewTodoHandler はTodoHandlerの新しいインスタンスを生成する
func NewTodoHandler(queries *dbsqlc.Queries, db *sql.DB) *TodoHandler {
	return &TodoHandler{
		queries: queries,
		db:      db,
	}
}

// stringToNullString は文字列をsql.NullStringに変換する
func ptrStringToNullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

// nullStringToString はsql.NullStringを文字列に変換する
func nullStringToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

// toTodoResponse はdb.Todoをmodel.TodoResponseに変換する
func toTodoResponse(t dbsqlc.Todo) model.TodoResponse {
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

// ListTodos はTodoのリストを取得する
func (h *TodoHandler) ListTodos(ctx context.Context, input *model.ListTodosInput) (*model.ListTodosOutput, error) {
	var todos []dbsqlc.Todo
	var err error

	if input.Completed {
		todos, err = h.queries.ListTodosByStatus(ctx, 1)
	} else {
		todos, err = h.queries.ListTodos(ctx)
	}

	if err != nil {
		slog.Warn("todoリストの取得に失敗", "err", err)
		return nil, huma.Error500InternalServerError("Todoリストの取得に失敗", err)
	}

	output := &model.ListTodosOutput{}
	output.Body.Todos = make([]model.TodoResponse, len(todos))
	for i, t := range todos {
		output.Body.Todos[i] = toTodoResponse(t)
	}

	return output, nil
}

// GetTodo は指定されたIDのTodoを取得する
func (h *TodoHandler) GetTodo(ctx context.Context, input *model.GetTodoInput) (*model.GetTodoOutput, error) {
	todo, err := h.queries.GetTodo(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Todo IDが見つかりません", "id", input.ID, "err", err)
			return nil, huma.Error404NotFound(fmt.Sprintf("Todo IDが見つかりません: %d", input.ID))
		}
		slog.Warn("Todoの取得に失敗", "err", err)
		return nil, huma.Error500InternalServerError("Todo取得に失敗", err)
	}

	return &model.GetTodoOutput{Body: toTodoResponse(todo)}, nil
}

// CreateTodo は新しいTodoを作成する
func (h *TodoHandler) CreateTodo(ctx context.Context, input *model.CreateTodoInput) (*model.CreateTodoOutput, error) {
	description := ptrStringToNullString(input.Body.Description)

	todo, err := h.queries.CreateTodo(ctx, dbsqlc.CreateTodoParams{
		Title:       input.Body.Title,
		Description: description,
		Completed:   0,
	})
	if err != nil {
		slog.Warn("Todo作成に失敗", "err", err)
		return nil, huma.Error500InternalServerError("Todo作成に失敗", err)
	}

	return &model.CreateTodoOutput{Body: toTodoResponse(todo)}, nil
}

// UpdateTodo は指定されたIDのTodoを更新する
func (h *TodoHandler) UpdateTodo(ctx context.Context, input *model.UpdateTodoInput) (*model.UpdateTodoOutput, error) {
	_, err := h.queries.GetTodo(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma.Error404NotFound(fmt.Sprintf("IDが見つかりませんでした: %d", input.ID))
		}
		return nil, huma.Error500InternalServerError("Todoの取得に失敗", err)
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Warn("トランザクション開始に失敗", "err", err)
		return nil, huma.Error500InternalServerError("トランザクション開始に失敗", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := h.queries.WithTx(tx)

	var completed int64
	if input.Body.Completed {
		completed = 1
	}

	description := ptrStringToNullString(input.Body.Description)

	todo, err := qtx.UpdateTodo(ctx, dbsqlc.UpdateTodoParams{
		ID:          input.ID,
		Title:       input.Body.Title,
		Description: description,
		Completed:   completed,
	})
	if err != nil {
		slog.Warn("Todo更新に失敗", "err", err)
		return nil, huma.Error500InternalServerError("Todo更新に失敗", err)
	}

	if err := tx.Commit(); err != nil {
		slog.Warn("トランザクションのコミットに失敗", "err", err)
		return nil, huma.Error500InternalServerError("トランザクションのコミットに失敗", err)
	}

	return &model.UpdateTodoOutput{Body: toTodoResponse(todo)}, nil
}

// DeleteTodo は指定されたIDのTodoを削除する
func (h *TodoHandler) DeleteTodo(ctx context.Context, input *model.DeleteTodoInput) (*model.DeleteTodoOutput, error) {
	_, err := h.queries.GetTodo(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma.Error404NotFound(fmt.Sprintf("IDが見つかりませんでした: %d", input.ID))
		}
		return nil, huma.Error500InternalServerError("Todoの取得に失敗", err)
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Warn("トランザクション開始に失敗", "err", err)
		return nil, huma.Error500InternalServerError("トランザクション開始に失敗", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := h.queries.WithTx(tx)

	if err := qtx.DeleteTodo(ctx, input.ID); err != nil {
		slog.Warn("Todo削除に失敗", "err", err)
		return nil, huma.Error500InternalServerError("Todo削除に失敗", err)
	}

	if err := tx.Commit(); err != nil {
		slog.Warn("トランザクションのコミットに失敗", "err", err)
		return nil, huma.Error500InternalServerError("トランザクションのコミットに失敗", err)
	}

	output := &model.DeleteTodoOutput{}
	output.Body.Message = "Todo deleted successfully"
	return output, nil
}

// ToggleTodo は指定されたIDのTodoの完了状態を切り替える
func (h *TodoHandler) ToggleTodo(ctx context.Context, input *model.ToggleTodoInput) (*model.ToggleTodoOutput, error) {
	_, err := h.queries.GetTodo(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma.Error404NotFound(fmt.Sprintf("IDが見つかりませんでした: %d", input.ID))
		}
		return nil, huma.Error500InternalServerError("Todoの取得に失敗", err)
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Warn("トランザクション開始に失敗", "err", err)
		return nil, huma.Error500InternalServerError("トランザクション開始に失敗", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := h.queries.WithTx(tx)

	todo, err := qtx.ToggleTodoCompleted(ctx, input.ID)
	if err != nil {
		slog.Warn("Todoのトグルに失敗", "err", err)
		return nil, huma.Error500InternalServerError("Todoのトグルに失敗", err)
	}

	if err := tx.Commit(); err != nil {
		slog.Warn("トランザクションのコミットに失敗", "err", err)
		return nil, huma.Error500InternalServerError("トランザクションのコミットに失敗", err)
	}

	return &model.ToggleTodoOutput{Body: toTodoResponse(todo)}, nil
}
