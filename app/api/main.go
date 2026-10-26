package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// #region main
func main() {
	// 接続先は環境変数から読む（ConfigMap と Secret で渡す）
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s"+
			" sslmode=disable",
		env("DB_HOST", "localhost"), env("DB_PORT", "5432"),
		env("DB_USER", "postgres"), env("DB_PASSWORD", ""),
		env("DB_NAME", "tododb"))
	store := &pgStore{}
	// DB を待たずに HTTP を起動し、接続は裏で続ける
	go store.connectLoop(dsn)

	srv := &http.Server{
		Addr:              ":" + env("PORT", "8080"),
		Handler:           newMux(store, prometheus.NewRegistry()),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

// #endregion main

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
