package controller

type reqOrderInsert struct {
	BuyerID         int64   `json:"buyer_id" validate:"required"`
	VenueID         int64   `json:"venue_id" validate:"required`
	ProductID       int64   `json:"product_id" validate:"required"`
	Quantity        int64   `json:"quantity" validate:"required"`
	TotalPrice      float64 `json:"total_price" validate:"required"`
	PaymentMethodID int64   `json:"payment_method_id" validate:"required"`
	PaymentFee      float64 `json:"payment_fee" validate:"required"`
}

type reqOrderUpdate struct {
	BuyerID         int64   `json:"buyer_id" validate:"required"`
	VenueID         int64   `json:"venue_id" validate:"required`
	ProductID       int64   `json:"product_id" validate:"required"`
	Quantity        int64   `json:"quantity" validate:"required"`
	TotalPrice      float64 `json:"total_price" validate:"required"`
	PaymentMethodID int64   `json:"payment_method_id" validate:"required"`
	PaymentFee      float64 `json:"payment_fee" validate:"required"`
	Status          int16   `json:"status"`
}
