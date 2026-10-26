package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// memStore はテスト用のメモリ上の Store。
type memStore struct {
	todos map[int]Todo
	next  int
}

func newMemStore() *memStore {
	return &memStore{todos: map[int]Todo{}, next: 1}
}

func (s *memStore) List(context.Context) ([]Todo, error) {
	out := []Todo{}
	for i := 1; i < s.next; i++ {
		if t, ok := s.todos[i]; ok {
			out = append(out, t)
		}
	}
	return out, nil
}

func (s *memStore) Create(
	_ context.Context, title string,
) (Todo, error) {
	t := Todo{ID: s.next, Title: title, CreatedAt: time.Now()}
	s.todos[t.ID] = t
	s.next++
	return t, nil
}

func (s *memStore) Update(
	_ context.Context, id int, p Patch,
) (Todo, error) {
	t, ok := s.todos[id]
	if !ok {
		return t, ErrNotFound
	}
	if p.Title != nil {
		t.Title = *p.Title
	}
	if p.Completed != nil {
		t.Completed = *p.Completed
	}
	s.todos[id] = t
	return t, nil
}

func (s *memStore) Ready(context.Context) error { return nil }

func (s *memStore) Delete(_ context.Context, id int) error {
	if _, ok := s.todos[id]; !ok {
		return ErrNotFound
	}
	delete(s.todos, id)
	return nil
}

func do(t *testing.T, h http.Handler, method, path, body string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

func TestCRUD(t *testing.T) {
	mux := newMux(newMemStore(), prometheus.NewRegistry())
	steps := []struct {
		method, path, body string
		want               int
	}{
		{"POST", "/api/todos", `{"title":"牛乳を買う"}`, 201},
		{"POST", "/api/todos", `{"title":""}`, 400},
		{"GET", "/api/todos", "", 200},
		{"PUT", "/api/todos/1", `{"completed":true}`, 200},
		{"PUT", "/api/todos/9", `{"completed":true}`, 404},
		{"DELETE", "/api/todos/1", "", 204},
		{"DELETE", "/api/todos/1", "", 404},
		{"GET", "/healthz", "", 200},
		{"GET", "/readyz", "", 200},
	}
	for _, s := range steps {
		if got := do(t, mux, s.method, s.path, s.body); got != s.want {
			t.Errorf("%s %s = %d, want %d",
				s.method, s.path, got, s.want)
		}
	}
}

// TestReadyz は DB に接続する前の /readyz が 503 を返し、
// /healthz は 200 のままであることを確かめる。
func TestReadyz(t *testing.T) {
	mux := newMux(&pgStore{}, prometheus.NewRegistry())
	if got := do(t, mux, "GET", "/readyz", ""); got != 503 {
		t.Errorf("/readyz = %d, want 503", got)
	}
	if got := do(t, mux, "GET", "/healthz", ""); got != 200 {
		t.Errorf("/healthz = %d, want 200", got)
	}
}

// TestMetrics は API の操作の回数が todo_api_requests_total に
// ルートのパターンで記録されることを確かめる。
func TestMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	mux := newMux(newMemStore(), reg)
	do(t, mux, "POST", "/api/todos", `{"title":"a"}`)
	do(t, mux, "PUT", "/api/todos/1", `{"completed":true}`)
	do(t, mux, "GET", "/metrics", "")

	m := newMetricsForRead(t, reg)
	got := testutil.ToFloat64(m.WithLabelValues(
		"PUT", "/api/todos/{id}", "200"))
	if got != 1 {
		t.Fatalf("PUT の回数 = %v, want 1", got)
	}
	if n := testutil.CollectAndCount(m); n != 2 {
		t.Fatalf("ラベルの組 = %d, want 2（/metrics は数えない）", n)
	}
}

// newMetricsForRead は登録済みの Counter を取り出す。
func newMetricsForRead(
	t *testing.T, reg *prometheus.Registry,
) *prometheus.CounterVec {
	t.Helper()
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "todo_api_requests_total",
		Help: "API が処理したリクエストの数",
	}, []string{"method", "path", "status"})
	err := reg.Register(c)
	var are prometheus.AlreadyRegisteredError
	if err == nil || !asAlready(err, &are) {
		t.Fatalf("登録済みの Counter が見つからない: %v", err)
	}
	return are.ExistingCollector.(*prometheus.CounterVec)
}

func asAlready(err error, are *prometheus.AlreadyRegisteredError) bool {
	a, ok := err.(prometheus.AlreadyRegisteredError)
	if ok {
		*are = a
	}
	return ok
}
