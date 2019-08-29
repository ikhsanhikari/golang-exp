package controller

type reqEmail struct {
	VenueID int64 `json:"venueId"`
}

type reqInvoice struct {
	OrderID int64 `json:"orderID"`
}
