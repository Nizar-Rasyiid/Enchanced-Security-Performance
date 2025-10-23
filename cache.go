package main

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client

func initRedis(addr string) {
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Gagal koneksi Redis: %v", err)
	}
}
