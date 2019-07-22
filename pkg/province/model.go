package province

type Province struct {
	ProvinceID int64  `db:"province_id"`
	Province   string `db:"province"`
	CountryID  string `db:"country_id"`
	AppID      string `db:"app_id"`
	ProjectID  int64  `db:"project_id"`
}

type Provinces []Province
