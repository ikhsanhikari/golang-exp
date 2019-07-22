package view

type ProvinceAttributes struct {
	ProvinceID int64          `json:"province_id"`
	Province   string         `json:"province"`
	CountryID  string         `json:"country_id"`
	AppID      string         `json:"app_id"`
	ProjectID  int64          `json:"project_id"`
	Cities     []DataResponse `json:"cities"`
}
