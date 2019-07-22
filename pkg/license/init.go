package license

import (
	"context"
	"log"
	"time"

	auditTrail "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/audit_trail"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

// Init is used to initialize license package
func Init(db *sqlx.DB, redis *redis.Pool, auditTrail auditTrail.ICore) ICore {
	examineDBHealth(db)
	return &core{
		db:         db,
		redis:      redis,
		auditTrail: auditTrail,
	}
}

func examineDBHealth(db *sqlx.DB) {
	if db == nil {
		log.Fatalf("Failed to initialize license. db object cannot be nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.PingContext(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize license. cannot pinging to db. err: %s", err)
	}
}
