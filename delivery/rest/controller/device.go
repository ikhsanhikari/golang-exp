package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/device"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := c.device.Select(10)
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
				Name:      device.Name,
				Info:      device.Info,
				Price:     device.Price,
				Status:    device.Status,
				ProjectID: device.ProjectID,
				CreatedAt: device.CreatedAt,
				UpdatedAt: device.UpdatedAt,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteDevice(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteDevice] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.device.Get(10, id)
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

	err = c.device.Delete(10, id)
	if err != nil {
		c.reporter.Errorf("[handleDeleteDevice] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete device", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostDevice(w http.ResponseWriter, r *http.Request) {
	var params reqDevice
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostDevice] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	device := device.Device{
		Name:      params.Name,
		Info:      params.Info,
		Price:     params.Price,
		ProjectID: 10,
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
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchDevice] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqDevice
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchDevice] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.device.Get(10, id)
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

	device := device.Device{
		ID:        id,
		Name:      params.Name,
		Info:      params.Info,
		Price:     params.Price,
		ProjectID: 10,
	}
	err = c.device.Update(&device)
	if err != nil {
		c.reporter.Errorf("[handlePatchDevice] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update device", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, device, http.StatusOK)
}
