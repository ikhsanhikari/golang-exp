package order_matrix

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

//OrderMatrix is model for mla_order_matrix in db
type OrderMatrix struct {
	ID             int64     `db:"id"`
	VenueTypeID    int64     `db:"venue_type_id"`
	Capacity       *int64    `db:"capacity"`
	AgingID        int64     `db:"aging_id"`
	DeviceID       int64     `db:"device_id"`
	RoomID         *int64    `db:"room_id"`
	ProductID      int64     `db:"product_id"`
	InstallationID int64     `db:"installation_id"`
	Status         int16     `db:"status"`
	CreatedAt      time.Time `db:"created_at"`
	CreatedBy      string    `db:"created_by"`
	UpdatedAt      time.Time `db:"updated_at"`
	LastUpdateBy   string    `db:"last_update_by"`
	DeletedAt      null.Time `db:"deleted_at"`
	ProjectID      int64     `db:"project_id"`
}

//OrderMatrices is list of order matrix
type OrderMatrices []OrderMatrix

//OrderMatrixDetail is model for get all matrix with their name
type OrderMatrixDetail struct {
	ID               int64     `db:"id"`
	VenueTypeID      int64     `db:"venue_type_id"`
	VenueTypeName    string    `db:"venue_type_name"`
	Capacity         int64     `db:"capacity"`
	AgingID          int64     `db:"aging_id"`
	AgingName        string    `db:"aging_name"`
	DeviceID         int64     `db:"device_id"`
	DeviceName       string    `db:"device_name"`
	RoomID           int64     `db:"room_id"`
	RoomName         string    `db:"room_name"`
	ProductID        int64     `db:"product_id"`
	ProductName      string    `db:"product_name"`
	InstallationID   int64     `db:"installation_id"`
	InstallationName string    `db:"installation_name"`
	Status           int16     `db:"status"`
	CreatedAt        time.Time `db:"created_at"`
	CreatedBy        string    `db:"created_by"`
	UpdatedAt        time.Time `db:"updated_at"`
	LastUpdateBy     string    `db:"last_update_by"`
	DeletedAt        null.Time `db:"deleted_at"`
	ProjectID        int64     `db:"project_id"`
}

//OrderMatrixDetails is list of OrderMatrixDetail
type OrderMatrixDetails []OrderMatrixDetail
