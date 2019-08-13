package controller

type reqSubscription struct {
	PackageDuration int64  `json:"packageDuration"`
	BoxSerialNumber string `json:"boxSerialNumber"`
	SmartCardNumber string `json:"smartCardNumber"`
	Status          int8   `json:"status"`
	ProjectID       int64  `json:"projectId"`
	CreatedBy       string `json:"createdBy"`
	LastUpdateBy    string `json:"lastUpdateBy"`
}
