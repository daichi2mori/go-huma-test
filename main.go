package main

import (
	"database/sql"
	"fmt"
	"go-huma-test/db"
	"go-huma-test/handler"
	"go-huma-test/model"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	_ "github.com/mattn/go-sqlite3"
)

func initDB(dbPath string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		completed INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TRIGGER IF NOT EXISTS update_todos_updated_at
		AFTER UPDATE ON todos
		FOR EACH ROW
	BEGIN
		UPDATE todos SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
	END;
	`

	if _, err := sqlDB.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return sqlDB, nil
}

func main() {
	sqlDB, err := initDB("./todos.db")
	if err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	handler := handler.NewTodoHandler(queries)

	cli := humacli.New(func(h humacli.Hooks, o *model.Options) {
		mux := http.NewServeMux()

		config := huma.DefaultConfig("Todo API", "1.0.0")
		config.Info.Description = "SQLite + sqlc + Humaを使ったシンプルなTodo API"
		api := humago.New(mux, config)

		huma.Register(api, huma.Operation{
			OperationID: "list-todos",
			Method:      http.MethodGet,
			Path:        "/todos",
			Summary:     "Todo一覧取得",
			Description: "すべてのTodoを取得",
			Tags:        []string{"todos"},
		}, handler.ListTodos)

		huma.Register(api, huma.Operation{
			OperationID: "get-todo",
			Method:      http.MethodGet,
			Path:        "/todos/{id}",
			Summary:     "Todo取得",
			Description: "指定したIDのTodoを取得します。",
			Tags:        []string{"todos"},
		}, handler.GetTodo)

		huma.Register(api, huma.Operation{
			OperationID:   "create-todo",
			Method:        http.MethodPost,
			Path:          "/todos",
			Summary:       "Todo作成",
			Description:   "新しいTodoを作成します。",
			Tags:          []string{"todos"},
			DefaultStatus: http.StatusCreated,
		}, handler.CreateTodo)

		huma.Register(api, huma.Operation{
			OperationID: "update-todo",
			Method:      http.MethodPut,
			Path:        "/todos/{id}",
			Summary:     "Todo更新",
			Description: "指定したIDのTodoを更新します。",
			Tags:        []string{"todos"},
		}, handler.UpdateTodo)

		huma.Register(api, huma.Operation{
			OperationID: "delete-todo",
			Method:      http.MethodDelete,
			Path:        "/todos/{id}",
			Summary:     "Todo削除",
			Description: "指定したIDのTodoを削除します。",
			Tags:        []string{"todos"},
		}, handler.DeleteTodo)

		huma.Register(api, huma.Operation{
			OperationID: "toggle-todo",
			Method:      http.MethodPost,
			Path:        "/todos/{id}/toggle",
			Summary:     "Todo完了状態切り替え",
			Description: "指定したIDのTodoの完了状態を切り替えます。",
			Tags:        []string{"todos"},
		}, handler.ToggleTodo)

		h.OnStart(func() {
			addr := fmt.Sprintf(":%d", o.Port)
			log.Printf("🚀 Todo API Server starting on http://localhost%s", addr)
			log.Printf("📚 API Documentation: http://localhost%s/docs", addr)
			log.Printf("📚 Get OpenAPI File: http://localhost%s/openapi.yaml", addr)
			if err := http.ListenAndServe(addr, mux); err != nil {
				log.Fatalf("Server failed: %v", err)
			}
		})

	})

	cli.Run()
}
