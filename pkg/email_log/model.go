package email_log

import (
	"time"

	null "gopkg.in/guregu/null.v3"
)

type EmailLog struct {
	ID           int64     `db:"id"`
	SenderUID    string    `db:"sender_uid"`
	OrderID      int64     `db:"order_id"`
	VenueID      int64     `db:"venue_id"`
	CompanyID    int64     `db:"company_id"`
	To           string    `db:"to_email"`
	EmailType    string    `db:"email_type"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	DeletedAt    null.Time `db:"deleted_at"`
	Status       int64     `db:"status"`
	ProjectID    int64     `db:"project_id"`
	CreatedBy    string    `db:"created_by"`
	LastUpdateBy string    `db:"last_update_by"`
}

type EmailLogs []EmailLog
