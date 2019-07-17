package controller

import (
	"time"

	"gopkg.in/guregu/null.v3"
)
type orderData struct {
	OrderNumber     string    
	BuyerID         string    
	VenueID         int64     
	DeviceID        int64     
	ProductID       int64     
	InstallationID  int64     
	Quantity        int64     
	AgingID         int64     
	RoomID          int64     
	RoomQuantity    int64     
	TotalPrice      string   
	PaymentMethodID int64     
	PaymentFee      float64   
	Status          int16     
	CreatedAt       string
	CreatedBy       string    
	UpdatedAt       time.Time 
	LastUpdateBy    string    
	DeletedAt       null.Time 
	PendingAt       null.Time 
	PaidAt          null.Time 
	FailedAt        null.Time 
	ProjectID       int64     
	Email           string    
	Description     string 
	ProductName		string
	ProductPrice	string
	DeviceName      string 
	InstallationName string 
	AgingName		 string 
	RoomName		 string 
}

type orderDatass []orderData

