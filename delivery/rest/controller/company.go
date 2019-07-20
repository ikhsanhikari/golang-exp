package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/company"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	//auth "git.sstv.io/lib/go/go-auth-api.git/authpassport"
)

func (c *Controller) handleGetAllCompanies(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetAllCompanies] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleGetAllCompanies] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	companies, err := c.company.Select(10, fmt.Sprintf("%v", userID))
	if err != nil {
		c.reporter.Errorf("[handleGetAllCompanies] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get company", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(companies))
	for _, company := range companies {
		res = append(res, view.DataResponse{
			Type: "company",
			ID:   company.ID,
			Attributes: view.CompanyAttributes{
				ID:           company.ID,
				Name:         company.Name,
				Address:      company.Address,
				City:         company.City,
				Province:     company.Province,
				Zip:          company.Zip,
				Email:        company.Email,
				Npwp:         company.Npwp,
				CreatedAt:    company.CreatedAt,
				UpdatedAt:    company.UpdatedAt,
				DeletedAt:    company.DeletedAt,
				ProjectID:    company.ProjectID,
				CreatedBy:    company.CreatedBy,
				LastUpdateBy: company.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetCompanyByID(w http.ResponseWriter, r *http.Request) {
	var (
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleGetCompanyByID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetCompanyByID] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleGetCompanyByID] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	company, err := c.company.Get(id, 10, fmt.Sprintf("%v", userID))
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
		Type: "company",
		ID:   company.ID,
		Attributes: view.CompanyAttributes{
			ID:           company.ID,
			Name:         company.Name,
			Address:      company.Address,
			City:         company.City,
			Province:     company.Province,
			Zip:          company.Zip,
			Email:        company.Email,
			Npwp:         company.Npwp,
			CreatedAt:    company.CreatedAt,
			UpdatedAt:    company.UpdatedAt,
			DeletedAt:    company.DeletedAt,
			ProjectID:    company.ProjectID,
			CreatedBy:    company.CreatedBy,
			LastUpdateBy: company.LastUpdateBy,
		},
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (c *Controller) handleDeleteCompany(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteCompany] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteCompany] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleDeleteCompany] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	_, err = c.company.Get(id, 10, fmt.Sprintf("%v", userID))
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteCompany] Company not found, err: %s", err.Error())
		view.RenderJSONError(w, "Company not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteCompany] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get company", http.StatusInternalServerError)
		return
	}

	err = c.company.Delete(id, 10, fmt.Sprintf("%v", userID))
	if err != nil {
		c.reporter.Errorf("[handleDeleteCompany] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete Company", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostCompany(w http.ResponseWriter, r *http.Request) {
	var params reqCompany
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostCompany] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostCompany] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}

	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handlePostCompany] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	company := company.Company{
		ID:        params.ID,
		Name:      params.Name,
		Address:   params.Address,
		City:      params.City,
		Province:  params.Province,
		Zip:       params.Zip,
		Email:     params.Email,
		Npwp:      params.Npwp,
		CreatedBy: fmt.Sprintf("%v", userID),
	}

	err = c.company.Insert(&company)
	if err != nil {
		c.reporter.Infof("[handlePostCompany] error insert Company repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post Company", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, company, http.StatusOK)
}

func (c *Controller) handlePatchCompany(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchCompany] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePatchCompany] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handlePatchCompany] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	var params reqCompany
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchCompany] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.company.Get(id, 10, fmt.Sprintf("%v", userID))
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchCompany] Company not found, err: %s", err.Error())
		view.RenderJSONError(w, "Company not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchCompany] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Company", http.StatusInternalServerError)
		return
	}
	company := company.Company{
		ID:           id,
		Name:         params.Name,
		Address:      params.Address,
		City:         params.City,
		Province:     params.Province,
		Zip:          params.Zip,
		Email:        params.Email,
		Npwp:         params.Npwp,
		LastUpdateBy: fmt.Sprintf("%v", userID),
	}
	err = c.company.Update(&company)
	if err != nil {
		c.reporter.Errorf("[handlePatchCompany] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update Company", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, company, http.StatusOK)
}
