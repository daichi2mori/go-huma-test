// Package config はアプリケーションの設定を管理する。
// このパッケージはYAMLファイルから設定を読み込み、
// アプリケーション全体で使用する設定情報を提供する。
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config はアプリケーション全体の設定を表す構造体
type Config struct {
	Title       string         `env:"TITLE"`
	Version     string         `env:"VERSION"`
	Description string         `env:"DESCRIPTION"`
	Server      ServerConfig   `envPrefix:"SERVER_"`
	Database    DatabaseConfig `envPrefix:"DB_"`
}

// ServerConfig はサーバーの設定を表す構造体
type ServerConfig struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT"`
}

// DatabaseConfig はデータベースの設定を表す構造体
type DatabaseConfig struct {
	Path string `env:"PATH"`
}

// Load は環境変数から設定値を取得
func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("環境変数のパースに失敗: %w", err)
	}

	return &cfg, nil
}
