package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/aging"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handlePostAging(w http.ResponseWriter, r *http.Request) {
	var (
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		params reqAging
	)

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePostAging] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	aging := aging.Aging{
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		ProjectID:   10,
	}

	err = c.aging.Insert(&aging)
	if err != nil {
		c.reporter.Errorf("[handlePostAging] failed post aging, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post aging", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID:   aging.ID,
		Type: "aging",
		Attributes: view.AgingAttributesResponse{
			Name:        aging.Name,
			Description: aging.Description,
			Price:       aging.Price,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handlePatchAging(w http.ResponseWriter, r *http.Request) {
	var (
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		params  reqAging
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handlePatchAging] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.aging.Get(id, 10)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchAging] aging not found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchAging] Failed get aging, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get aging", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePatchAging] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	aging := aging.Aging{
		ID:          id,
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		ProjectID:   10,
	}

	err = c.aging.Update(&aging)
	if err != nil {
		c.reporter.Errorf("[handlePatchAging] failed update aging, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update aging", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID:   aging.ID,
		Type: "aging",
		Attributes: view.AgingAttributesResponse{
			Name:        aging.Name,
			Description: aging.Description,
			Price:       aging.Price,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteAging(w http.ResponseWriter, r *http.Request) {
	var (
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleDeleteAging] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.aging.Get(id, 10)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteAging] aging not found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteAging] failed get aging, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get aging", http.StatusInternalServerError)
		return
	}

	err = c.aging.Delete(id, 10)
	if err != nil {
		c.reporter.Errorf("[handleDeleteAging] failed delete aging, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete aging", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID: id,
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllAgings(w http.ResponseWriter, r *http.Request) {
	var (
	// project, _ = authpassport.GetProject(r)
	// pid        = project.ID
	)

	agings, err := c.aging.Select(10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllAging] aging not found, err: %s", err.Error())
		view.RenderJSONError(w, "Orders not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetAllAging] failed get aging, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get aging", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponseOrder, 0, len(agings))
	for _, aging := range agings {
		res = append(res, view.DataResponseOrder{
			ID:   aging.ID,
			Type: "aging",
			Attributes: view.AgingAttributes{
				Name:        aging.Name,
				Description: aging.Description,
				Price:       aging.Price,
				Status:      aging.Status,
				CreatedAt:   aging.CreatedAt,
				UpdatedAt:   aging.UpdatedAt,
				DeletedAt:   aging.DeletedAt,
				ProjectID:   aging.ProjectID,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}
