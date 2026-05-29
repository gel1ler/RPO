// Package main RPO REST API.
//
// @title           RPO Authorization API
// @version         1.0
// @description     REST API сервера авторизации платежей транспортными картами.
// @BasePath        /api/v1
// @schemes         https
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT из POST /auth/login: «Bearer &lt;token&gt;» или только &lt;token&gt;
package main

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"rpo/internal/auth"
	"rpo/internal/httpapi"
	"rpo/internal/store"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	dbPath := os.Getenv("APP_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/app.db"
	}

	if err := os.MkdirAll(dirOf(dbPath), 0o755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	dsn := "file:" + dbPath + "?_foreign_keys=on"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("sql open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	adminLogin := os.Getenv("APP_ADMIN_LOGIN")
	adminPassword := os.Getenv("APP_ADMIN_PASSWORD")
	if adminLogin != "" && adminPassword != "" {
		hash, err := auth.HashPassword(adminPassword)
		if err != nil {
			log.Fatalf("admin password hash: %v", err)
		}
		if _, err := (store.Users{DB: db}).EnsureAdmin(context.Background(), adminLogin, hash); err != nil {
			log.Fatalf("ensure admin: %v", err)
		}
		log.Printf("admin ensured: %s", adminLogin)
	}

	terminalSerial := os.Getenv("APP_TERMINAL_SERIAL")
	terminalSecret := os.Getenv("APP_TERMINAL_API_SECRET")
	if terminalSerial != "" && terminalSecret != "" {
		hash, err := auth.HashPassword(terminalSecret)
		if err != nil {
			log.Fatalf("terminal api secret hash: %v", err)
		}
		if _, err := (store.Terminals{DB: db}).EnsureTerminal(context.Background(), terminalSerial, hash); err != nil {
			log.Fatalf("ensure terminal: %v", err)
		}
		log.Printf("terminal ensured: %s", terminalSerial)
	}

	jwtSecret := os.Getenv("APP_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
	}

	api := httpapi.Server{
		DB: db,
		JWT: auth.JWT{
			Secret: []byte(jwtSecret),
			Issuer: "rpo",
			TTL:    24 * time.Hour,
		},
	}

	addr := os.Getenv("APP_HTTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	log.Printf("listening %s", addr)
	log.Fatal(http.ListenAndServe(addr, api.Router()))
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
