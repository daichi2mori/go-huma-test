// Package db はテーブル初期化用の SQL スキーマを読み込むパッケージです。
package db

import _ "embed"

// Schema は schema.sql の内容を文字列として保持します。
//
//go:embed schema.sql
var Schema string
