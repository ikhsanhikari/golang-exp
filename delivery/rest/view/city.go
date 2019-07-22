package view

type CityAttributes struct {
	ProvinceID string `json:"province_id"`
	City       string `json:"city"`
	AppID      string `json:"app_id"`
	ProjectID  int64  `json:"project_id"`
}
