package product

import (
	"context"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

// Init is used to initialize product package
func Init(db *sqlx.DB, redis *redis.Pool) ICore {
	examineDBHealth(db)
	return &core{
		db: db,
	}
}

func examineDBHealth(db *sqlx.DB) {
	if db == nil {
		log.Fatalf("Failed to initialize product. db object cannot be nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.PingContext(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize product. cannot pinging to db. err: %s", err)
	}
}
