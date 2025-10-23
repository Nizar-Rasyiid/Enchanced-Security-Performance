package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Initialize CIA security framework
	InitSecurityConfig()

	// konfigurasi: kalau mau konek DB/Redis, ubah DSN/addr; biar aman kita toleran jika tidak tersedia
	dsn := "" // contoh: "postgres://user:password@localhost:5432/appdb?sslmode=disable"
	redisAddr := "localhost:6379"

	var db *sql.DB
	if dsn != "" {
		db = openDB(dsn)
		defer func() {
			if db != nil {
				_ = db.Close()
			}
		}()
	}

	initRedis(redisAddr) // jika Redis tidak tersedia, hanya log warning

	r := setupRouter(db)

	// ensure certs exist (generate self-signed for dev if missing)
	certDir := "certs"
	_ = os.MkdirAll(certDir, 0755)
	certFile := filepath.Join(certDir, "server.crt")
	keyFile := filepath.Join(certDir, "server.key")
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Println("Sertifikat tidak ditemukan; membuat self-signed untuk pengujian...")
		if err := generateSelfSignedCert(certFile, keyFile); err != nil {
			log.Fatalf("Gagal membuat sertifikat: %v", err)
		}
	}

	srv := newSecureServer(":8443", r)

	log.Println("Server jalan di https://localhost:8443")
	if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	waitForShutdown(srv)
}
