package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/admin"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllAdmins(w http.ResponseWriter, r *http.Request) {
	admins, err := c.admin.Select(10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllAdmins] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Admins", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(admins))
	for _, admin := range admins {
		res = append(res, view.DataResponse{
			Type: "admins",
			ID:   admin.ID,
			Attributes: view.AdminAttributes{
				UserID:       admin.UserID,
				Status:       admin.Status,
				ProjectID:    admin.ProjectID,
				CreatedAt:    admin.CreatedAt,
				UpdatedAt:    admin.UpdatedAt,
				CreatedBy:    admin.CreatedBy,
				LastUpdateBy: admin.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllAdminsByUserID(w http.ResponseWriter, r *http.Request) {
	id := router.GetParam(r, "userId")

	admins, err := c.admin.SelectByUserID(10, id)
	if err != nil {
		c.reporter.Errorf("[handleGetAllAdminsByUserID] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Admins", http.StatusInternalServerError)
		return
	}
	status := false
	if len(admins) > 0 {
		status = true
		view.RenderJSONData(w, status, http.StatusOK)
		return
	}
	view.RenderJSONData(w, status, http.StatusOK)
}

func (c *Controller) handleDeleteAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteAdmin] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	adminParam, err := c.admin.Get(10, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteAdmin] admin not found, err: %s", err.Error())
		view.RenderJSONError(w, "admin not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteAdmin] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get admin", http.StatusInternalServerError)
		return
	}

	err = c.admin.Delete(10, id, adminParam.UserID)
	if err != nil {
		c.reporter.Errorf("[handleDeleteAdmin] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete admin", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostAdmin(w http.ResponseWriter, r *http.Request) {
	var params reqAdmin
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostAdmin] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	admin := admin.Admin{
		UserID:    params.UserID,
		ProjectID: 10,
		CreatedBy: params.CreatedBy,
	}

	err = c.admin.Insert(&admin)
	if err != nil {
		c.reporter.Infof("[handlePostAdmin] error insert admin repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post admin", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, admin, http.StatusOK)
}

func (c *Controller) handlePatchAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchAdmin] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqAdmin
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchAdmin] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	adminParam, err := c.admin.Get(10, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchAdmin] admin not found, err: %s", err.Error())
		view.RenderJSONError(w, "Admin not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchAdmin] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get admin", http.StatusInternalServerError)
		return
	}

	admin := admin.Admin{
		ID:           id,
		UserID:       params.UserID,
		ProjectID:    10,
		LastUpdateBy: params.LastUpdateBy,
	}
	err = c.admin.Update(&admin, adminParam.UserID)
	if err != nil {
		c.reporter.Errorf("[handlePatchAdmin] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update admin", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, admin, http.StatusOK)
}

func (c *Controller) handleAdminsCheck(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleAdminsCheck] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}

	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleAdminsCheck] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	_, isExist := c.admin.Check(fmt.Sprintf("%v", userID))
	if isExist == sql.ErrNoRows {
		c.reporter.Errorf("[handleAdminsCheck] user is not exist")
		view.RenderJSONError(w, "user is not exist", http.StatusUnauthorized)
		return
	}

	view.RenderJSON(w, nil, http.StatusOK)
}
