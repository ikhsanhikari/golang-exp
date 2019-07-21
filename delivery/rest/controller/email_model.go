package controller

type reqEmail struct {
	To      string `json:"to"`
	OrderID int64  `json:"orderId"`
}
