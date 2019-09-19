package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	//auth "git.sstv.io/lib/go/go-auth-api.git/authpassport"
)

func (c *Controller) handleGetAllCities(w http.ResponseWriter, r *http.Request) {
	cities, err := c.city.Select(c.projectID)
	if err != nil {
		c.reporter.Errorf("[handleGetAllCities] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get city", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(cities))
	for _, cities := range cities {
		res = append(res, view.DataResponse{
			Type: "city",
			ID:   cities.CityID,
			Attributes: view.CityAttributes{
				ProvinceID: cities.ProvinceID,
				City:       cities.City,
				AppID:      cities.AppID,
				ProjectID:  cities.ProjectID,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetCityByID(w http.ResponseWriter, r *http.Request) {
	var (
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleGetCompanyByID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	cities, err := c.city.Get(id, c.projectID)
	if err != nil {
		c.reporter.Errorf("[handleGetCompanyByID] company not found, err: %s", err.Error())
		view.RenderJSONError(w, "Company not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetCompanyByID] failed get company, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get company", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		Type: "city",
		ID:   cities.CityID,
		Attributes: view.CityAttributes{
			ProvinceID: cities.ProvinceID,
			City:       cities.City,
			AppID:      cities.AppID,
			ProjectID:  cities.ProjectID,
		},
	}
	view.RenderJSONData(w, res, http.StatusOK)
}
