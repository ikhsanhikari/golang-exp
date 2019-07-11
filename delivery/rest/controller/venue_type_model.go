package controller

type reqVenueType struct {
	Id         						int64     `db:"id"`
	Name    						string    `db:"name"`
	Description  					string    `db:"description"`
	Capacity    					int64     `db:"capacity"`
	PricingGroupID   				int64     `db:"pricingGroupId"`
	CommercialTypeID        		int64     `db:"commercialTypeId"`
}
