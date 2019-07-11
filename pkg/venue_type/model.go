package venue_type

import (
	"time"
	null "gopkg.in/guregu/null.v3"
)

type VenueType struct {
	Id         						int64     `db:"id"`
	Name    						string    `db:"name"`
	Description  					string    `db:"description"`
	Capacity    					int64     `db:"capacity"`
	PricingGroupID   				int64     `db:"pricing_group_id"`
	CommercialTypeID        		int64     `db:"commercial_type_id"`
	CreatedAt  						time.Time `db:"created_at"`
	UpdatedAt  						time.Time `db:"updated_at"`
	DeletedAt  						null.Time `db:"deleted_at"`
	Status     						int64     `db:"status"`
	ProjectID   					int64     `db:"project_id"`
}

type VenueTypes []VenueType
