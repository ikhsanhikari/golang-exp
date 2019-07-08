package controller

import (
	"database/sql"
	"net/http"
	"strconv"


	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/commercial_type"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllcommercialTypes(w http.ResponseWriter, r *http.Request) {
	commercialTypes, err := c.commercialType.Select(10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllcommercialTypes] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get commercialType", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(commercialTypes))
	for _, commercialType := range commercialTypes {
		res = append(res, view.DataResponse{
			Type: "commercialType",
			ID:   commercialType.ID,
			Attributes: view.CommercialTypeAttributes{
				ID				:  commercialType.ID,
				Name			:  commercialType.Name,
				Description		:  commercialType.Description,
				CreatedAt		:  commercialType.CreatedAt,
				UpdatedAt		:  commercialType.UpdatedAt,
				DeletedAt		:  commercialType.DeletedAt,
				ProjectID		:  commercialType.ProjectID,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (c *Controller) handleDeletecommercialType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeletecommercialType] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.commercialType.Get(id,10)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeletecommercialType] commercialType not found, err: %s", err.Error())
		view.RenderJSONError(w, "commercialType not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeletecommercialType] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get commercialType", http.StatusInternalServerError)
		return
	}

	err = c.commercialType.Delete(id,10)
	if err != nil {
		c.reporter.Errorf("[handleDeletecommercialType] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete commercialType", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostcommercialType(w http.ResponseWriter, r *http.Request) {
	var params reqCommercialType
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostcommercialType] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	
	commercialType := commercial_type.CommercialType{
		ID				:  params.ID,
		Name			:  params.Name,
		Description		:  params.Description,
	}

	err = c.commercialType.Insert(&commercialType)
	if err != nil {
		c.reporter.Infof("[handlePostcommercialType] error insert Commercial_type repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post commercialType", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, commercialType, http.StatusOK)
}

func (c *Controller) handlePatchcommercialType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchcommercialType] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqCommercialType
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchcommercialType] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.commercialType.Get(id,10)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchcommercialType] commercialType not found, err: %s", err.Error())
		view.RenderJSONError(w, "commercialType not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchcommercialType] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get commercialType", http.StatusInternalServerError)
		return
	}
	commercialType := commercial_type.CommercialType{
		ID				:  id,
		Name			:  params.Name,
		Description		:  params.Description,
	}
	err = c.commercialType.Update(&commercialType)
	if err != nil {
		c.reporter.Errorf("[handlePatchcommercialType] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update commercialType", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, commercialType, http.StatusOK)
}
 