package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// openDB mencoba koneksi DB jika dsn tidak kosong.
// jika dsn kosong -> kembalikan nil (tidak fatal).
func openDB(dsn string) *sql.DB {
	if dsn == "" {
		log.Println("[DB] DSN kosong, melewatkan koneksi DB")
		return nil
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Gagal koneksi DB: %v", err)
	}
	// konfigurasi pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	return db
}
