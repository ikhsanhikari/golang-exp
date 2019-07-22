package audit_trail

import (
	"time"
)

type AuditTrail struct {
	ID        int64     `db:"id"`
	UserID    string    `db:"user_id"`
	Query     string    `db:"query_executed"`
	TableName string    `db:"table_name"`
	ProjectID int64     `db:"project_id"`
	Timestamp time.Time `db:"timestamp"`
}

type AuditTrails []AuditTrail
