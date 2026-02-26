// Package main はTodo管理APIのエントリーポイントを提供する。
// このパッケージはHumaフレームワークを使用してREST APIサーバーを起動し、
// SQLiteデータベースと連携してTodoの管理機能を提供する。
package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-huma-test/config"
	"go-huma-test/handler"
	"go-huma-test/repository"
	"go-huma-test/usecase"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	dbpkg "go-huma-test/db"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
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

	// DI
	todoRepo := repository.NewTodoRepository(sqlDB)
	todoUC := usecase.NewTodoUseCase(todoRepo)
	todoHandler := handler.NewTodoHandler(todoUC)

	mux := http.NewServeMux()
	humaCfg := huma.DefaultConfig(cfg.Title, cfg.Version)
	humaCfg.Info.Description = cfg.Description
	humaCfg.CreateHooks = []func(huma.Config) huma.Config{}
	api := humago.New(mux, humaCfg)
	todoHandler.RegisterRoutes(api)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,  // ヘッダ読み取り制限
		ReadTimeout:       15 * time.Second, // 全体の読み取り制限
		WriteTimeout:      15 * time.Second, // レスポンス書き込み制限
		IdleTimeout:       60 * time.Second, // keep-alive制御
	}

	go func() {
		fmt.Printf("🚀 Todo API Server starting on http://%s\n", addr)
		fmt.Printf("📚 API Documentation: http://%s/docs\n", addr)
		fmt.Printf("📚 Get OpenAPI File: http://%s/openapi.yaml\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("サーバー起動に失敗", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("サーバーのシャットダウンに失敗", "err", err)
		os.Exit(1)
	}
	slog.Info("サーバーは正常にシャットダウンされました")
}
