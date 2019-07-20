package controller

import (
	"database/sql"
	"net/http"
	"strconv"


	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/company"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	//auth "git.sstv.io/lib/go/go-auth-api.git/authpassport"
)

func (c *Controller) handleGetAllCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := c.company.Select(10)
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
				ID				:  company.ID,
				Name			:  company.Name,
				Address			:  company.Address,
				City			:  company.City,
				Province		:  company.Province,
				Zip				:  company.Zip,
				Npwp			:  company.Npwp,
				CreatedAt		:  company.CreatedAt,
				UpdatedAt		:  company.UpdatedAt,
				DeletedAt		:  company.DeletedAt,
				ProjectID		:  company.ProjectID,
				CreatedBy		:  company.CreatedBy,
				LastUpdateBy	:  company.LastUpdateBy,
			},
		})
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

	_, err = c.company.Get(id,10)
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

	err = c.company.Delete(id,10)
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
	
	company := company.Company{
		ID				:  params.ID,
		Name			:  params.Name,
		Address			:  params.Address,
		City			:  params.City,
		Province		:  params.Province,
		Zip				:  params.Zip,
		Npwp			:  params.Npwp,
		CreatedBy		:  params.CreatedBy,
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

	var params reqCompany
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchCompany] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.company.Get(id,10)
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
		ID				:  id,
		Name			:  params.Name,
		Address			:  params.Address,
		City			:  params.City,
		Province		:  params.Province,
		Zip				:  params.Zip,
		Npwp			:  params.Npwp,
		LastUpdateBy	:  params.LastUpdateBy,
	}
	err = c.company.Update(&company)
	if err != nil {
		c.reporter.Errorf("[handlePatchCompany] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update Company", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, company, http.StatusOK)
}
 