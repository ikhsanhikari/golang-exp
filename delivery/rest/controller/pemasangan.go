package controller

import (
	"database/sql"
	"net/http"
	"strconv"


	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/pemasangan"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllPemasangans(w http.ResponseWriter, r *http.Request) {
	pemasangans, err := c.pemasangan.Select(10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllPemasangans] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get pemasangan", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(pemasangans))
	for _, pemasangan := range pemasangans {
		res = append(res, view.DataResponse{
			Type: "pemasangan",
			ID:   pemasangan.ID,
			Attributes: view.PemasanganAttributes{
				ID				:  pemasangan.ID,
				Description		:  pemasangan.Description,
				Price			:  pemasangan.Price,
				DeviceID		:  pemasangan.DeviceID,
				CreatedAt		:  pemasangan.CreatedAt,
				UpdatedAt		:  pemasangan.UpdatedAt,
				DeletedAt		:  pemasangan.DeletedAt,
				ProjectID		:  pemasangan.ProjectID,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (c *Controller) handleDeletePemasangan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeletePemasangan] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.pemasangan.Get(id,10)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeletePemasangan] Pemasangan not found, err: %s", err.Error())
		view.RenderJSONError(w, "Pemasangan not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeletePemasangan] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Pemasangan", http.StatusInternalServerError)
		return
	}

	err = c.pemasangan.Delete(id,10)
	if err != nil {
		c.reporter.Errorf("[handleDeletePemasangan] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete Pemasangan", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostPemasangan(w http.ResponseWriter, r *http.Request) {
	var params reqPemasangan
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostPemasangan] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	
	pemasangan := pemasangan.Pemasangan{
		ID				:  params.ID,
		Description		:  params.Description,
		Price			:  params.Price,
		DeviceID		:  params.DeviceID,
	}

	err = c.pemasangan.Insert(&pemasangan)
	if err != nil {
		c.reporter.Infof("[handlePostPemasangan] error insert Pemasangan repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post Pemasangan", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, pemasangan, http.StatusOK)
}

func (c *Controller) handlePatchPemasangan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchPemasangan] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqPemasangan
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchPemasangan] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.pemasangan.Get(id,10)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchPemasangan] Pemasangan not found, err: %s", err.Error())
		view.RenderJSONError(w, "Pemasangan not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchPemasangan] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Pemasangan", http.StatusInternalServerError)
		return
	}
	pemasangan := pemasangan.Pemasangan{
		ID				:  id,
		Description		:  params.Description,
		Price			:  params.Price,
		DeviceID		:  params.DeviceID,
	}
	err = c.pemasangan.Update(&pemasangan)
	if err != nil {
		c.reporter.Errorf("[handlePatchPemasangan] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update Pemasangan", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, pemasangan, http.StatusOK)
}
 