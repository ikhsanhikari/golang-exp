package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/license"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	auth "git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	"git.sstv.io/lib/go/gojunkyard.git/util"
)

func (c *Controller) handleGetAllLicenses(w http.ResponseWriter, r *http.Request) {
	var (
		pid = int64(10)
	)
	licenses, err := c.license.Select(pid)
	if err != nil {
		c.reporter.Errorf("[handleGetAllLicenses] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Licenses", http.StatusInternalServerError)
		return
	}
	res := make([]view.DataResponse, 0, len(licenses))
	for _, license := range licenses {
		res = append(res, view.DataResponse{
			Type: "licenses",
			ID:   license.ID,
			Attributes: view.LicenseAttributes{
				LicenseNumber: license.LicenseNumber,
				OrderID:       license.OrderID,
				LicenseStatus: license.LicenseStatus,
				ActiveDate:    license.ActiveDate,
				ExpiredDate:   license.ExpiredDate,
				Status:        license.Status,
				ProjectID:     license.ProjectID,
				CreatedAt:     license.CreatedAt,
				UpdatedAt:     license.UpdatedAt,
				CreatedBy:     license.CreatedBy,
				LastUpdateBy:  license.LastUpdateBy,
				BuyerID:       license.BuyerID,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetLicensesByBuyerID(w http.ResponseWriter, r *http.Request) {
	var (
		pid = int64(10)
	)
	buyerID := router.GetParam(r, "buyer_id")
	licenses, err := c.license.GetByBuyerId(pid, buyerID)
	if err != nil {
		c.reporter.Errorf("[handleGetLicensesByBuyerId] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Licenses by buyer id", http.StatusInternalServerError)
		return
	}
	res := make([]view.DataResponse, 0, len(licenses))
	for _, license := range licenses {
		res = append(res, view.DataResponse{
			Type: "licenses",
			ID:   license.ID,
			Attributes: view.LicenseAttributes{
				LicenseNumber: license.LicenseNumber,
				OrderID:       license.OrderID,
				LicenseStatus: license.LicenseStatus,
				ActiveDate:    license.ActiveDate,
				ExpiredDate:   license.ExpiredDate,
				Status:        license.Status,
				ProjectID:     license.ProjectID,
				CreatedAt:     license.CreatedAt,
				UpdatedAt:     license.UpdatedAt,
				CreatedBy:     license.CreatedBy,
				LastUpdateBy:  license.LastUpdateBy,
				BuyerID:       license.BuyerID,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteLicense(w http.ResponseWriter, r *http.Request) {
	var (
		pid     = int64(10)
		params  reqDeleteLicense
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)
	if err != nil {
		c.reporter.Warningf("[handleDeleteLicense] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	licenseParam, err := c.license.Get(pid, id)
	buyerID := licenseParam.BuyerID
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteLicense] license not found, err: %s", err.Error())
		view.RenderJSONError(w, "license not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteLicense] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get license", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handleDeleteLicense] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	//check user id
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteLicense] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handleDeleteLicense] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}
	err = c.license.Delete(pid, id, buyerID, licenseParam.LicenseNumber, isAdmin, userID.(string))
	if err != nil {
		c.reporter.Errorf("[handleDeleteLicense] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete license", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostLicense(w http.ResponseWriter, r *http.Request) {
	var (
		pid    = int64(10)
		params reqLicense
	)

	user, ok := auth.GetUser(r)
	if !ok {
		c.reporter.Warningf("[handlePostLicense] [GetUser] Failed get user")
		view.RenderJSONError(w, "Failed get user from token", http.StatusBadRequest)
		return
	}
	uid, ok := user["sub"]
	if !ok {
		c.reporter.Warningf("[handlePostLicense] user[sub] Failed get id user because nil")
	}
	err := form.Bind(&params, r)

	if uid != nil {
		uidStr := fmt.Sprintf("%v", uid)
		params.CreatedBy = uidStr
	}

	if err != nil {
		c.reporter.Warningf("[handlePostLicense] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	licenseNumberUUID := util.GenerateUUID()

	license := license.License{
		LicenseNumber: licenseNumberUUID,
		OrderID:       params.OrderID,
		LicenseStatus: params.LicenseStatus,
		ActiveDate:    params.ActiveDate,
		ExpiredDate:   params.ExpiredDate,
		ProjectID:     pid,
		CreatedBy:     params.CreatedBy,
		BuyerID:       params.BuyerID,
	}

	err = c.license.Insert(&license)
	if err != nil {
		c.reporter.Infof("[handlePostLicense] error insert license repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post license", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, license, http.StatusOK)
}

func (c *Controller) handlePatchLicense(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchLicense] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqLicense

	user, ok := auth.GetUser(r)
	if !ok {
		c.reporter.Warningf("[handlePatchLicense][GetUser] Failed get user")
		view.RenderJSONError(w, "Failed get user from token", http.StatusBadRequest)
		return
	}
	uid, ok := user["sub"]
	if !ok {
		c.reporter.Warningf("[handlePatchLicense] user[sub] Failed get id user because nil")
	}

	uid = "newUID"
	err = form.Bind(&params, r)

	if uid != nil {
		uidStr := fmt.Sprintf("%v", uid)
		params.LastUpdateBy = uidStr
	}

	if err != nil {
		c.reporter.Warningf("[handlePatchLicense] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	licenseParam, err := c.license.Get(10, id)
	buyerID := licenseParam.BuyerID
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchLicense] license not found, err: %s", err.Error())
		view.RenderJSONError(w, "License not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchLicense] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get license", http.StatusInternalServerError)
		return
	}

	license := license.License{
		ID:            id,
		OrderID:       params.OrderID,
		LicenseStatus: params.LicenseStatus,
		ActiveDate:    params.ActiveDate,
		ExpiredDate:   params.ExpiredDate,
		ProjectID:     10,
		LastUpdateBy:  params.LastUpdateBy,
		BuyerID:       params.BuyerID,
	}
	err = c.license.Update(&license, buyerID)
	if err != nil {
		c.reporter.Errorf("[handlePatchLicense] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update license", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, license, http.StatusOK)
}
