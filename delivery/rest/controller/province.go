package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	//auth "git.sstv.io/lib/go/go-auth-api.git/authpassport"
)

func (c *Controller) handleGetAllProvinces(w http.ResponseWriter, r *http.Request) {

	provinces, err := c.province.Select(c.projectID)
	if err != nil {
		c.reporter.Errorf("[handleGetAllCities] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get province", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(provinces))
	for _, provinces := range provinces {
		res = append(res, view.DataResponse{
			Type: "province",
			ID:   provinces.ProvinceID,
			Attributes: view.ProvinceAttributes{
				ProvinceID: provinces.ProvinceID,
				Province:   provinces.Province,
				CountryID:  provinces.CountryID,
				AppID:      provinces.AppID,
				ProjectID:  provinces.ProjectID,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetProvincesByID(w http.ResponseWriter, r *http.Request) {
	var (
		_id        = router.GetParam(r, "id")
		id, err    = strconv.ParseInt(_id, 10, 64)
		citiesResp = make([]view.DataResponse, 0, 1)
	)
	if err != nil {
		c.reporter.Errorf("[handleGetCompanyByID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	provinces, err := c.province.Get(id, c.projectID)
	if err != nil && err == sql.ErrNoRows {
		c.reporter.Errorf("[handleGetProvincesByID] province not found, err: %s", err.Error())
		view.RenderJSONError(w, "Company not found", http.StatusNotFound)
		return
	}
	if err != nil {
		c.reporter.Errorf("[handleGetCompanyByID] failed get province, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get company", http.StatusInternalServerError)
		return
	}
	getParam := r.URL.Query()
	if getParam.Get("relationship") == "true" {
		cities, err := c.city.SelectByProvince(provinces.ProvinceID, c.projectID)
		if err != nil && err == sql.ErrNoRows {
			c.reporter.Errorf("[handleSelectByProvince] select city in province not found, err: %s", err.Error())
			view.RenderJSONError(w, "City not found in province", http.StatusNotFound)
			return
		}
		if err != nil {
			c.reporter.Errorf("[handleSelectByProvince] failed  select city in province, err: %s", err.Error())
			view.RenderJSONError(w, "Failed select city in province", http.StatusInternalServerError)
			return
		}

		for _, c := range cities {
			citiesResp = append(citiesResp, view.DataResponse{
				ID:   c.CityID,
				Type: "city",
				Attributes: view.CityAttributes{
					ProvinceID: c.ProvinceID,
					City:       c.City,
					AppID:      c.AppID,
					ProjectID:  c.ProjectID,
				},
			})
		}
	}
	res := view.DataResponse{
		Type: "province",
		ID:   provinces.ProvinceID,
		Attributes: view.ProvinceAttributes{
			ProvinceID: provinces.ProvinceID,
			Province:   provinces.Province,
			CountryID:  provinces.CountryID,
			AppID:      provinces.AppID,
			ProjectID:  provinces.ProjectID,
			Cities:     citiesResp,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}
