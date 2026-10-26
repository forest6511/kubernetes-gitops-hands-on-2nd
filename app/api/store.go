package main

import (
	"context"
	"errors"
	"time"
)

// #region store
// Todo は1件のタスク。JSON の名前は frontend と合わせる。
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// Store は Todo の保存先。本番は PostgreSQL、
// テストはメモリ上の実装に差し替える。
type Store interface {
	List(ctx context.Context) ([]Todo, error)
	Create(ctx context.Context, title string) (Todo, error)
	Update(ctx context.Context, id int, p Patch) (Todo, error)
	Delete(ctx context.Context, id int) error
}

// Patch は更新する項目。nil の項目は変えない。
type Patch struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

// #endregion store

var (
	// ErrNotFound は指定した ID の Todo が無いことを表す。
	ErrNotFound = errors.New("not found")
	// ErrNotReady は DB にまだ接続していないことを表す。
	ErrNotReady = errors.New("database not ready")
)
