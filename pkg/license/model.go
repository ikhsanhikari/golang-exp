package license

import "time"
import "gopkg.in/guregu/null.v3"

type License struct {
	ID            int64     `db:"id"`
	LicenseNumber string    `db:"license_number"`
	OrderID       int64     `db:"venue_id"`
	LicenseStatus int8      `db:"license_status"`
	ActiveDate    time.Time `db:"active_date"`
	ExpiredDate   time.Time `db:"expired_date"`
	Status        int8      `db:"status"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
	DeletedAt     null.Time `db:"deleted_at"`
	ProjectID     int64     `db:"project_id"`
	CreatedBy     string    `db:"created_by"`
	LastUpdateBy  string    `db:"last_update_by"`
	BuyerID       string    `db:"buyer_id"`
}

type Licenses []License
