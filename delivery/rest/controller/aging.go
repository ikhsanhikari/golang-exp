package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/aging"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handlePostAging(w http.ResponseWriter, r *http.Request) {
	var (
		projectID = int64(10)
		params    reqInsertAging
	)

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePostAging] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostAging] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		if params.CreatedBy == "" {
			c.reporter.Errorf("[handlePostAging] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.CreatedBy
	}

	aging := aging.Aging{
		Name:         params.Name,
		Description:  params.Description,
		Price:        params.Price,
		ProjectID:    projectID,
		CreatedBy:    userID.(string),
		LastUpdateBy: userID.(string),
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
		Attributes: view.AgingAttributes{
			Name:         aging.Name,
			Description:  aging.Description,
			Price:        aging.Price,
			Status:       aging.Status,
			CreatedAt:    aging.CreatedAt,
			CreatedBy:    aging.CreatedBy,
			UpdatedAt:    aging.UpdatedAt,
			LastUpdateBy: aging.LastUpdateBy,
			DeletedAt:    aging.DeletedAt,
			ProjectID:    aging.ProjectID,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handlePatchAging(w http.ResponseWriter, r *http.Request) {
	var (
		projectID = int64(10)
		params    reqUpdateAging
		_id       = router.GetParam(r, "id")
		id, err   = strconv.ParseInt(_id, 10, 64)
		isAdmin   = false
	)
	if err != nil {
		c.reporter.Errorf("[handlePatchAging] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	getAging, err := c.aging.Get(id, projectID)
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

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePatchAging] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		if params.LastUpdateBy == "" {
			c.reporter.Errorf("[handlePatchAging] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.LastUpdateBy
		isAdmin = true
	}

	aging := aging.Aging{
		ID:           id,
		Name:         params.Name,
		Description:  params.Description,
		Price:        params.Price,
		ProjectID:    projectID,
		CreatedBy:    getAging.CreatedBy,
		LastUpdateBy: userID.(string),
	}

	err = c.aging.Update(&aging, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePatchAging] failed update aging, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update aging", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID:   aging.ID,
		Type: "aging",
		Attributes: view.AgingAttributes{
			Name:         aging.Name,
			Description:  aging.Description,
			Price:        aging.Price,
			Status:       getAging.Status,
			CreatedAt:    getAging.CreatedAt,
			CreatedBy:    aging.CreatedBy,
			UpdatedAt:    aging.UpdatedAt,
			LastUpdateBy: aging.LastUpdateBy,
			DeletedAt:    getAging.DeletedAt,
			ProjectID:    aging.ProjectID,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteAging(w http.ResponseWriter, r *http.Request) {
	var (
		_id       = router.GetParam(r, "id")
		id, err   = strconv.ParseInt(_id, 10, 64)
		params    reqDeleteAging
		projectID = int64(10)
		isAdmin   = false
	)
	if err != nil {
		c.reporter.Errorf("[handleDeleteAging] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteAging] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handlePostAging] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	_, err = c.aging.Get(id, projectID)
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

	err = c.aging.Delete(id, projectID, userID.(string), isAdmin)
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
		projectID = int64(10)
	)

	agings, err := c.aging.Select(projectID)
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
				Name:         aging.Name,
				Description:  aging.Description,
				Price:        aging.Price,
				Status:       aging.Status,
				CreatedAt:    aging.CreatedAt,
				CreatedBy:    aging.CreatedBy,
				UpdatedAt:    aging.UpdatedAt,
				LastUpdateBy: aging.LastUpdateBy,
				DeletedAt:    aging.DeletedAt,
				ProjectID:    aging.ProjectID,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}
