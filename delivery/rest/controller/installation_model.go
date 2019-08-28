package controller

type reqInstallation struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	DeviceID     int64   `json:"deviceId"`
	CreatedBy    string  `json:"createdBy"`
	LastUpdateBy string  `json:"lastUpdateBy"`
}

type reqDeleteInstallation struct {
	UserID string `json:"userID"`
}
