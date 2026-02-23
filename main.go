// Package main はTodo管理APIのエントリーポイントを提供する。
// このパッケージはHumaフレームワークを使用してREST APIサーバーを起動し、
// SQLiteデータベースと連携してTodoの管理機能を提供する。
package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-huma-test/config"
	dbsqlc "go-huma-test/db/sqlc"
	"go-huma-test/handler"
	"go-huma-test/model"
	"log/slog"
	"net/http"
	"os"
	"time"

	dbpkg "go-huma-test/db"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	_ "github.com/mattn/go-sqlite3"
)

func initDB(dbPath string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("データベース接続に失敗: %w", err)
	}

	params := []string{
		"PRAGMA busy_timeout = 5000;", // ロックされている場合最大5秒待つ
		"PRAGMA journal_mode = WAL;",  // 読み取りは複数同時に可能だが書き込みは１つだけ。SQLiteをWebAPIで使用する場合はほぼ必須
		"PRAGMA foreign_keys = ON;",   // 外部キー制約を有効化（将来のために）
	}
	for _, p := range params {
		if _, err := sqlDB.Exec(p); err != nil {
			return nil, err
		}
	}

	sqlDB.SetMaxOpenConns(1) // 同時に開ける最大コネクション数
	sqlDB.SetMaxIdleConns(1) // アイドル状態のコネクション数

	if _, err := sqlDB.Exec(dbpkg.Schema); err != nil {
		return nil, fmt.Errorf("データベース初期化スキーマの実行失敗: %w", err)
	}

	slog.Info("データベース接続に成功")

	return sqlDB, nil
}

func LoggingMiddleware(ctx huma.Context, next func(huma.Context)) {
	fmt.Printf("[%s] %s\n", ctx.Method(), ctx.URL().Path)
	next(ctx)
}

func AuthMiddleware(ctx huma.Context, next func(huma.Context)) {
	// 認証チェック
	token := ctx.Header("Authorization")
	if token == "" {
		slog.Warn("Authorizationが設定されていません")
		if err := huma.WriteErr(huma.NewAPI(huma.Config{}, nil), ctx, http.StatusUnauthorized, "Authorization header required"); err != nil {
			slog.Warn("エラーレスポンスの書き込みに失敗", "err", err)
		}
		return
	}

	next(ctx)
}

func main() {
	// ロガー初期化
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	})))

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		slog.Error("設定の読み込みに失敗", "err", err)
		os.Exit(1)
	}

	slog.Info("設定を読み込みました",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"db_path", cfg.Database.Path,
	)

	// マイグレーション
	sqlDB, err := initDB(cfg.Database.Path)
	if err != nil {
		slog.Error("データベース初期化に失敗", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			slog.Error("データベースの終了に失敗", "err", err)
		}
	}()

	queries, err := dbsqlc.Prepare(context.Background(), sqlDB)
	if err != nil {
		slog.Error("データベースのPrepareに失敗", "err", err)
		os.Exit(1)
	}
	handler := handler.NewTodoHandler(queries, sqlDB)

	cli := humacli.New(func(h humacli.Hooks, o *model.Options) {
		mux := http.NewServeMux()

		config := huma.DefaultConfig(cfg.Title, cfg.Version)
		config.Info.Description = cfg.Description
		config.CreateHooks = []func(huma.Config) huma.Config{}
		api := humago.New(mux, config)

		// ミドルウェア設定
		api.UseMiddleware(LoggingMiddleware)
		api.UseMiddleware(AuthMiddleware)

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

		srv := &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,  // ヘッダ読み取り制限
			ReadTimeout:       15 * time.Second, // 全体の読み取り制限
			WriteTimeout:      15 * time.Second, // レスポンス書き込み制限
			IdleTimeout:       60 * time.Second, // keep-alive制御
		}

		h.OnStart(func() {
			slog.Info("サーバー起動開始...")
			addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
			fmt.Printf("🚀 Todo API Server starting on http://%s\n", addr)
			fmt.Printf("📚 API Documentation: http://%s/docs\n", addr)
			fmt.Printf("📚 Get OpenAPI File: http://%s/openapi.yaml\n", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("サーバー起動に失敗", "err", err)
				os.Exit(1)
			}
		})

		h.OnStop(func() {
			slog.Info("Shutting down server...")
			slog.Info("サーバーのシャットダウン開始...")

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				slog.Error("サーバーのシャットダウンに失敗", "err", err)
				os.Exit(1)
			}

			slog.Info("サーバーは正常にシャットダウンされました")
		})
	})

	cli.Run()
}
