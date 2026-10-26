package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const createTable = `CREATE TABLE IF NOT EXISTS todos (
	id SERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	completed BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

// pgStore は PostgreSQL に Todo を保存する。
// 接続できるまで ErrNotReady を返す。
type pgStore struct {
	mu sync.RWMutex
	db *sql.DB
}

// connectLoop は DB に接続できるまで1秒ごとに試す。
// DB の Pod より先に API の Pod が起動しても落ちないようにする。
func (s *pgStore) connectLoop(dsn string) {
	for i := 1; ; i++ {
		db, err := sql.Open("pgx", dsn)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			_, err = db.Exec(createTable)
		}
		if err == nil {
			s.mu.Lock()
			s.db = db
			s.mu.Unlock()
			log.Println("connected to database")
			return
		}
		log.Printf("waiting for database (%d): %v", i, err)
		time.Sleep(time.Second)
	}
}

func (s *pgStore) conn() (*sql.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, ErrNotReady
	}
	return s.db, nil
}

func (s *pgStore) Ready(ctx context.Context) error {
	db, err := s.conn()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

func (s *pgStore) List(ctx context.Context) ([]Todo, error) {
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx,
		"SELECT id, title, completed, created_at"+
			" FROM todos ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	todos := []Todo{}
	for rows.Next() {
		var t Todo
		err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func (s *pgStore) Create(
	ctx context.Context, title string,
) (Todo, error) {
	var t Todo
	db, err := s.conn()
	if err != nil {
		return t, err
	}
	err = db.QueryRowContext(ctx,
		"INSERT INTO todos (title) VALUES ($1)"+
			" RETURNING id, title, completed, created_at",
		title).Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)
	return t, err
}

func (s *pgStore) Update(
	ctx context.Context, id int, p Patch,
) (Todo, error) {
	var t Todo
	db, err := s.conn()
	if err != nil {
		return t, err
	}
	err = db.QueryRowContext(ctx,
		"UPDATE todos SET title = COALESCE($1, title),"+
			" completed = COALESCE($2, completed) WHERE id = $3"+
			" RETURNING id, title, completed, created_at",
		p.Title, p.Completed, id,
	).Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

func (s *pgStore) Delete(ctx context.Context, id int) error {
	db, err := s.conn()
	if err != nil {
		return err
	}
	res, err := db.ExecContext(ctx,
		"DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
