package audit_trail

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// Init is used to initialize installation package
func Init(db *sqlx.DB, pid int64) ICore {
	examineDBHealth(db)
	return &core{
		db:  db,
		pid: pid,
	}
}

func examineDBHealth(db *sqlx.DB) {
	if db == nil {
		log.Fatalf("Failed to initialize audit trail. db object cannot be nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.PingContext(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize audit trail. cannot pinging to db. err: %s", err)
	}
}
