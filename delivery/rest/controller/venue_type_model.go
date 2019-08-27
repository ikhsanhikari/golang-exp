package controller

type reqVenueType struct {
	Id               int64  `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Capacity         int64  `json:"capacity"`
	PricingGroupID   int64  `json:"pricingGroupId"`
	CommercialTypeID int64  `json:"commercialTypeId"`
	CreatedBy        string `json:"createdBy"`
	LastUpdateBy     string `json:"lastUpdateBy"`
}

type reqDeleteVenueType struct {
	UserID string `json:"userID"`
}
