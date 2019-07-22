package city

type City struct {
	CityID     int64  `db:"city_id"`
	ProvinceID string `db:"province_id"`
	City       string `db:"city"`
	AppID      string `db:"app_id"`
	ProjectID  int64  `db:"project_id"`
}

type Cities []City
