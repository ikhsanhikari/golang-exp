package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/device"

	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllDevices(w http.ResponseWriter, r *http.Request) {
	var (
		pid = int64(10)
	)

	devices, err := c.device.Select(pid)
	if err != nil {
		c.reporter.Errorf("[handleGetAllDevices] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Devices", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(devices))
	for _, device := range devices {
		res = append(res, view.DataResponse{
			Type: "devices",
			ID:   device.ID,
			Attributes: view.DeviceAttributes{
				Name:         device.Name,
				Info:         device.Info,
				Price:        device.Price,
				Status:       device.Status,
				ProjectID:    device.ProjectID,
				CreatedAt:    device.CreatedAt,
				UpdatedAt:    device.UpdatedAt,
				CreatedBy:    device.CreatedBy,
				LastUpdateBy: device.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteDevice(w http.ResponseWriter, r *http.Request) {
	var (
		pid     = int64(10)
		params  reqDeleteDevice
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)
	if err != nil {
		c.reporter.Warningf("[handleDeleteDevice] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.device.Get(pid, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteDevice] device not found, err: %s", err.Error())
		view.RenderJSONError(w, "device not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteDevice] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get device", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handleDeleteDevice] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	//check user id
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteDevice] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handleDeleteDevice] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	err = c.device.Delete(pid, id, isAdmin, userID.(string))
	if err != nil {
		c.reporter.Errorf("[handleDeleteDevice] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete device", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostDevice(w http.ResponseWriter, r *http.Request) {
	var (
		pid    = int64(10)
		params reqDevice
	)
	err := form.Bind(&params, r)

	//checking if userID nil, it will be request
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostDevice] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	var uid = ""
	if !ok {
		uid = params.CreatedBy
	} else {
		uid = fmt.Sprintf("%v", userID)
	}

	if err != nil {
		c.reporter.Warningf("[handlePostDevice] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	device := device.Device{
		Name:      params.Name,
		Info:      params.Info,
		Price:     params.Price,
		ProjectID: pid,
		CreatedBy: uid,
	}

	err = c.device.Insert(&device)
	if err != nil {
		c.reporter.Infof("[handlePostDevice] error insert device repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post device", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, device, http.StatusOK)
}

func (c *Controller) handlePatchDevice(w http.ResponseWriter, r *http.Request) {
	var (
		pid     = int64(10)
		params  reqDevice
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)
	if err != nil {
		c.reporter.Warningf("[handlePatchDevice] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchDevice] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.device.Get(pid, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchDevice] device not found, err: %s", err.Error())
		view.RenderJSONError(w, "Device not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchDevice] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get device", http.StatusInternalServerError)
		return
	}

	//check user id
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePatchRoom] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.LastUpdateBy == "" {
			c.reporter.Errorf("[handlePatchRoom] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.LastUpdateBy
		isAdmin = true
	}

	device := device.Device{
		ID:           id,
		Name:         params.Name,
		Info:         params.Info,
		Price:        params.Price,
		ProjectID:    pid,
		LastUpdateBy: userID.(string),
	}
	err = c.device.Update(&device, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePatchDevice] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update device", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, device, http.StatusOK)
}