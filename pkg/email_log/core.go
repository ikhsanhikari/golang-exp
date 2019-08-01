package email_log

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

// ICore is the interface
type ICore interface {
	Insert(emailLog *EmailLog) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

func (c *core) Insert(emailLog *EmailLog) (err error) {
	emailLog.CreatedAt = time.Now()
	emailLog.UpdatedAt = emailLog.CreatedAt
	emailLog.ProjectID = 10
	emailLog.Status = 1
	emailLog.LastUpdateBy = emailLog.CreatedBy

	_, err = c.db.NamedExec(`
		INSERT INTO mla_email_log (
			sender_uid,
			order_id,
			venue_id,
			company_id,
			to_email,
			email_type,
			created_at,
			updated_at,
			project_id,
			created_by,
			last_update_by,
			status
		) VALUES (
			:sender_uid,
			:order_id,
			:venue_id,
			:company_id,
			:to_email,
			:email_type,
			:created_at,
			:updated_at,
			:project_id,
			:created_by,
			:last_update_by,
			:status
		)
	`, emailLog)

	return
}
