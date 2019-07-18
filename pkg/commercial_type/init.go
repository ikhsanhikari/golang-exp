package commercial_type

import (
	"context"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

// Init is used to initialize installation package
func Init(db *sqlx.DB, redis *redis.Pool) ICore {
	examineDBHealth(db)
	return &core{
		db: db,
		redis: redis,
	}
}

func examineDBHealth(db *sqlx.DB) {
	if db == nil {
		log.Fatalf("Failed to initialize commercial type. db object cannot be nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.PingContext(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize commercial type. cannot pinging to db. err: %s", err)
	}
}
