package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/installation"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllInstallations(w http.ResponseWriter, r *http.Request) {
	installations, err := c.installation.Select(10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllInstallations] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get installation", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(installations))
	for _, installation := range installations {
		res = append(res, view.DataResponse{
			Type: "installation",
			ID:   installation.ID,
			Attributes: view.InstallationAttributes{
				ID:           installation.ID,
				Name:         installation.Name,
				Description:  installation.Description,
				Price:        installation.Price,
				DeviceID:     installation.DeviceID,
				CreatedAt:    installation.CreatedAt,
				UpdatedAt:    installation.UpdatedAt,
				DeletedAt:    installation.DeletedAt,
				ProjectID:    installation.ProjectID,
				CreatedBy:    installation.CreatedBy,
				LastUpdateBy: installation.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (c *Controller) handleDeleteInstallation(w http.ResponseWriter, r *http.Request) {
	var (
		params  reqDeleteInstallation
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)

	if err != nil {
		c.reporter.Warningf("[handleDeleteInstallation] id must be integer, err: %s", err.Error())
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
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handlePostAging] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	_, err = c.installation.Get(id, 10)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteInstallation] Installation not found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteInstallation] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Installation", http.StatusInternalServerError)
		return
	}

	err = c.installation.Delete(id, 10, userID.(string), isAdmin)
	if err != nil {
		c.reporter.Errorf("[handleDeleteInstallation] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete Installation", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostInstallation(w http.ResponseWriter, r *http.Request) {
	var params reqInstallation
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostInstallation] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	installation := installation.Installation{
		ID:          params.ID,
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		DeviceID:    params.DeviceID,
		CreatedBy:   params.CreatedBy,
	}

	err = c.installation.Insert(&installation)
	if err != nil {
		c.reporter.Infof("[handlePostInstallation] error insert Installation repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post Installation", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, installation, http.StatusOK)
}

func (c *Controller) handlePatchInstallation(w http.ResponseWriter, r *http.Request) {
	var (
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)
	if err != nil {
		c.reporter.Warningf("[handlePatchInstallation] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqInstallation
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchInstallation] form binding, err: %s", err.Error())
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

	_, err = c.installation.Get(id, 10)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchInstallation] Installation not found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchInstallation] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Installation", http.StatusInternalServerError)
		return
	}
	installation := installation.Installation{
		ID:           id,
		Name:         params.Name,
		Description:  params.Description,
		Price:        params.Price,
		DeviceID:     params.DeviceID,
		LastUpdateBy: userID.(string),
	}
	err = c.installation.Update(&installation, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePatchInstallation] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update Installation", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, installation, http.StatusOK)
}