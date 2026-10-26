package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// #region routes
// newMux はルーティングを組み立てる。API の4つだけを計測し、
// /healthz・/readyz・/metrics は計測しない。
func newMux(s Store, reg *prometheus.Registry) *http.ServeMux {
	m := newMetrics(reg)
	h := &handler{store: s}
	mux := http.NewServeMux()
	route := func(method, path string, f http.HandlerFunc) {
		mux.HandleFunc(method+" "+path, m.instrument(method, path, f))
	}
	route("GET", "/api/todos", h.list)
	route("POST", "/api/todos", h.create)
	route("PUT", "/api/todos/{id}", h.update)
	route("DELETE", "/api/todos/{id}", h.delete)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter,
		_ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", h.ready)
	mux.Handle("GET /metrics",
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	return mux
}

// #endregion routes

type handler struct{ store Store }

// ready は DB に接続して読み書きできるときだけ 200 を返す。
// /healthz はプロセスが応答するかだけを見る（livenessProbe 用）。
func (h *handler) ready(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ready(r.Context()); err != nil {
		writeErr(w, ErrNotReady)
		return
	}
	_, _ = w.Write([]byte("ok"))
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.List(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title string `json:"title"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Title == "" {
		writeJSON(w, http.StatusBadRequest,
			map[string]string{"error": "title is required"})
		return
	}
	t, err := h.store.Create(r.Context(), in.Title)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	var p Patch
	if err != nil || json.NewDecoder(r.Body).Decode(&p) != nil {
		writeJSON(w, http.StatusBadRequest,
			map[string]string{"error": "invalid request"})
		return
	}
	t, err := h.store.Update(r.Context(), id, p)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			map[string]string{"error": "invalid id"})
		return
	}
	if err := h.store.Delete(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErr はエラーの種類をステータスコードに直す。
func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeJSON(w, http.StatusNotFound,
			map[string]string{"error": "not found"})
	case errors.Is(err, ErrNotReady):
		writeJSON(w, http.StatusServiceUnavailable,
			map[string]string{"error": "database not ready"})
	default:
		log.Printf("internal error: %v", err)
		writeJSON(w, http.StatusInternalServerError,
			map[string]string{"error": "internal error"})
	}
}
