package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// #region metrics
// metrics は API の処理そのものを数える（/metrics の収集は数えない）。
// 名前とラベルは第4章の PromQL と同じにする。
type metrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "todo_api_requests_total",
			Help: "API が処理したリクエストの数",
		}, []string{"method", "path", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "todo_api_request_duration_seconds",
			Help:    "API の処理にかかった秒数",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
	}
	reg.MustRegister(m.requests, m.duration)
	return m
}

// #endregion metrics

// #region instrument
// instrument はハンドラーを包み、処理の回数と時間を記録する。
// path にはルートのパターン（/api/todos/{id}）を入れ、
// ID ごとにラベルが増えないようにする。
func (m *metrics) instrument(
	method, path string, h http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		h(rec, r)
		m.requests.WithLabelValues(
			method, path, strconv.Itoa(rec.status)).Inc()
		m.duration.WithLabelValues(method, path).
			Observe(time.Since(start).Seconds())
	}
}

// #endregion instrument

// statusRecorder は書き込まれたステータスコードを覚えておく。
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
