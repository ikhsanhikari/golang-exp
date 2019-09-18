package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order_matrix"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handlePostOrderMatrix(w http.ResponseWriter, r *http.Request) {
	var params reqOrderMatrix

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePostOrderMatrix] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.venueType.Get(c.projectID, params.VenueTypeID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderMatrix] Venue Type Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue Type Not Found", http.StatusNotFound)
		return
	}

	_, err = c.aging.Get(params.AgingID, c.projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderMatrix] Aging Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging Not Found", http.StatusNotFound)
		return
	}

	_, err = c.device.Get(c.projectID, params.DeviceID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderMatrix] Device Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Device Not Found", http.StatusNotFound)
		return
	}

	if params.RoomID != nil {
		_, err = c.room.Get(c.projectID, *params.RoomID)
		if err == sql.ErrNoRows {
			c.reporter.Errorf("[handlePostOrderMatrix] Room Not Found, err: %s", err.Error())
			view.RenderJSONError(w, "Room Not Found", http.StatusNotFound)
			return
		}
	}

	_, err = c.product.Get(c.projectID, params.ProductID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderMatrix] Product Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Product Not Found", http.StatusNotFound)
		return
	}

	_, err = c.installation.Get(params.InstallationID, c.projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderMatrix] Installation Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation Not Found", http.StatusNotFound)
		return
	}

	matrix := order_matrix.OrderMatrix{
		VenueTypeID:    params.VenueTypeID,
		Capacity:       params.Capacity,
		AgingID:        params.AgingID,
		DeviceID:       params.DeviceID,
		RoomID:         params.RoomID,
		ProductID:      params.ProductID,
		InstallationID: params.InstallationID,
		CreatedBy:      params.UserID,
		LastUpdateBy:   params.UserID,
		ProjectID:      c.projectID,
	}

	checker, err := c.orderMatrix.MatrixChecker(matrix)
	if err != nil {
		c.reporter.Errorf("[handlePostOrderMatrix] failed doing matrix checker, err: %s", err.Error())
		view.RenderJSONError(w, "Failed doing matrix checker", http.StatusInternalServerError)
		return
	}
	if checker.IsExists == 1 {
		c.reporter.Errorf("[handlePostOrderMatrix] Duplicate matrix")
		view.RenderJSONError(w, "Duplicate matrix", http.StatusConflict)
		return
	}

	err = c.orderMatrix.Insert(&matrix)
	if err != nil {
		c.reporter.Errorf("[handlePostOrderMatrix] failed post order matrix, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post order matrix", http.StatusInternalServerError)
		return
	}

	res := view.DataResponseOrderMatrix{
		ID:   matrix.ID,
		Type: "orderMatrix",
		Attributes: view.OrderMatrixAttributes{
			VenueTypeID:    matrix.VenueTypeID,
			Capacity:       matrix.Capacity,
			AgingID:        matrix.AgingID,
			DeviceID:       matrix.DeviceID,
			RoomID:         matrix.RoomID,
			ProductID:      matrix.ProductID,
			InstallationID: matrix.InstallationID,
			Status:         matrix.Status,
			CreatedAt:      matrix.CreatedAt,
			CreatedBy:      matrix.CreatedBy,
			UpdatedAt:      matrix.UpdatedAt,
			LastUpdateBy:   matrix.LastUpdateBy,
			DeletedAt:      matrix.DeletedAt,
			ProjectID:      matrix.ProjectID,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)

}

func (c *Controller) handlePatchOrderMatrix(w http.ResponseWriter, r *http.Request) {
	var (
		params  reqOrderMatrix
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrderMatrix] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrderMatrix] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	getMatrix, err := c.orderMatrix.Get(id, c.projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrderMatrix] Order Matrix Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Order Matrix Not Found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrderMatrix] Failed get order matrix, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order matrix", http.StatusInternalServerError)
		return
	}

	_, err = c.venueType.Get(c.projectID, params.VenueTypeID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrderMatrix] Venue Type Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue Type Not Found", http.StatusNotFound)
		return
	}

	_, err = c.aging.Get(params.AgingID, c.projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrderMatrix] Aging Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging Not Found", http.StatusNotFound)
		return
	}

	_, err = c.device.Get(c.projectID, params.DeviceID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrderMatrix] Device Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Device Not Found", http.StatusNotFound)
		return
	}

	if params.RoomID != nil {
		_, err = c.room.Get(c.projectID, *params.RoomID)
		if err == sql.ErrNoRows {
			c.reporter.Errorf("[handlePostOrderMatrix] Room Not Found, err: %s", err.Error())
			view.RenderJSONError(w, "Room Not Found", http.StatusNotFound)
			return
		}
	}

	_, err = c.product.Get(c.projectID, params.ProductID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrderMatrix] Product Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Product Not Found", http.StatusNotFound)
		return
	}

	_, err = c.installation.Get(params.InstallationID, c.projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrderMatrix] Installation Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation Not Found", http.StatusNotFound)
		return
	}

	matrix := order_matrix.OrderMatrix{
		ID:             id,
		VenueTypeID:    params.VenueTypeID,
		Capacity:       params.Capacity,
		AgingID:        params.AgingID,
		DeviceID:       params.DeviceID,
		RoomID:         params.RoomID,
		ProductID:      params.ProductID,
		InstallationID: params.InstallationID,
		LastUpdateBy:   params.UserID,
		ProjectID:      c.projectID,
	}

	checker, err := c.orderMatrix.MatrixChecker(matrix)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrderMatrix] failed doing matrix checker, err: %s", err.Error())
		view.RenderJSONError(w, "Failed doing matrix checker", http.StatusInternalServerError)
		return
	}
	if checker.IsExists == 1 {
		c.reporter.Errorf("[handlePatchOrderMatrix] Duplicate matrix")
		view.RenderJSONError(w, "Duplicate matrix", http.StatusConflict)
		return
	}

	err = c.orderMatrix.Update(&matrix)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrderMatrix] failed post order matrix, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post order matrix", http.StatusInternalServerError)
		return
	}

	res := view.DataResponseOrderMatrix{
		ID:   matrix.ID,
		Type: "orderMatrix",
		Attributes: view.OrderMatrixAttributes{
			VenueTypeID:    matrix.VenueTypeID,
			Capacity:       matrix.Capacity,
			AgingID:        matrix.AgingID,
			DeviceID:       matrix.DeviceID,
			RoomID:         matrix.RoomID,
			ProductID:      matrix.ProductID,
			InstallationID: matrix.InstallationID,
			Status:         getMatrix.Status,
			CreatedAt:      getMatrix.CreatedAt,
			CreatedBy:      getMatrix.CreatedBy,
			UpdatedAt:      matrix.UpdatedAt,
			LastUpdateBy:   matrix.LastUpdateBy,
			DeletedAt:      getMatrix.DeletedAt,
			ProjectID:      matrix.ProjectID,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteOrderMatrix(w http.ResponseWriter, r *http.Request) {
	var (
		params  reqDeleteMatrix
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleDeleteOrderMatrix] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handleDeleteOrderMatrix] user id not found, err: %s", err.Error())
		view.RenderJSONError(w, "User ID not found", http.StatusBadRequest)
		return
	}

	_, err = c.orderMatrix.Get(id, c.projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteOrderMatrix] Order Matrix Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Order Matrix Not Found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteOrderMatrix] Failed get order matrix, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order matrix", http.StatusInternalServerError)
		return
	}

	matrix := order_matrix.OrderMatrix{
		ID:           id,
		LastUpdateBy: params.UserID,
		ProjectID:    c.projectID,
	}

	err = c.orderMatrix.Delete(&matrix)
	if err != nil {
		c.reporter.Errorf("[handleDeleteOrderMatrix] failed delete order matrix, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete order matrix", http.StatusInternalServerError)
		return
	}

	res := view.DataResponseOrderMatrix{
		ID: id,
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllOrderMatrices(w http.ResponseWriter, r *http.Request) {
	matrices, err := c.orderMatrix.SelectDetails(c.projectID)
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetAllOrderMatrix] failed get all orders matrix, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get all order matrix", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponseOrderMatrix, 0, len(matrices))
	for _, matrix := range matrices {
		res = append(res, view.DataResponseOrderMatrix{
			ID:   matrix.ID,
			Type: "orderMatrix",
			Attributes: view.OrderMatrixDetailAttributes{
				VenueTypeID:      matrix.VenueTypeID,
				VenueTypeName:    matrix.VenueTypeName,
				Capacity:         matrix.Capacity,
				AgingID:          matrix.AgingID,
				AgingName:        matrix.AgingName,
				DeviceID:         matrix.DeviceID,
				DeviceName:       matrix.DeviceName,
				RoomID:           matrix.RoomID,
				RoomName:         matrix.RoomName,
				ProductID:        matrix.ProductID,
				ProductName:      matrix.ProductName,
				InstallationID:   matrix.InstallationID,
				InstallationName: matrix.InstallationName,
				Status:           matrix.Status,
				CreatedAt:        matrix.CreatedAt,
				CreatedBy:        matrix.CreatedBy,
				UpdatedAt:        matrix.UpdatedAt,
				LastUpdateBy:     matrix.LastUpdateBy,
				DeletedAt:        matrix.DeletedAt,
				ProjectID:        matrix.ProjectID,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetOrderMatrixByID(w http.ResponseWriter, r *http.Request) {
	var (
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleGetOrderMatrixByID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	matrix, err := c.orderMatrix.GetDetails(id, c.projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleGetOrderMatrixByID] Order Matrix Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Order Matrix Not Found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetOrderMatrixByID] Failed get order matrix, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order matrix", http.StatusInternalServerError)
		return
	}

	res := view.DataResponseOrderMatrix{
		ID:   matrix.ID,
		Type: "orderMatrix",
		Attributes: view.OrderMatrixDetailAttributes{
			VenueTypeID:      matrix.VenueTypeID,
			VenueTypeName:    matrix.VenueTypeName,
			Capacity:         matrix.Capacity,
			AgingID:          matrix.AgingID,
			AgingName:        matrix.AgingName,
			DeviceID:         matrix.DeviceID,
			DeviceName:       matrix.DeviceName,
			RoomID:           matrix.RoomID,
			RoomName:         matrix.RoomName,
			ProductID:        matrix.ProductID,
			ProductName:      matrix.ProductName,
			InstallationID:   matrix.InstallationID,
			InstallationName: matrix.InstallationName,
			Status:           matrix.Status,
			CreatedAt:        matrix.CreatedAt,
			CreatedBy:        matrix.CreatedBy,
			UpdatedAt:        matrix.UpdatedAt,
			LastUpdateBy:     matrix.LastUpdateBy,
			DeletedAt:        matrix.DeletedAt,
			ProjectID:        matrix.ProjectID,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetVenueTypesFromMatrix(w http.ResponseWriter, r *http.Request) {
	venueTypes, err := c.orderMatrix.SelectVenueTypes(c.projectID)
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetVenueTypesFromMatrix] failed get all venue types, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get all venue types", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponseOrderMatrix, 0, len(venueTypes))
	for _, venueType := range venueTypes {
		res = append(res, view.DataResponseOrderMatrix{
			Type: "orderMatrix",
			Attributes: view.SummaryVenueTypeAttributes{
				VenueTypeID:   venueType.VenueTypeID,
				VenueTypeName: venueType.VenueTypeName,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetCapacitiesFromMatrix(w http.ResponseWriter, r *http.Request) {
	var params reqSumCapacities
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handleGetCapacitiesFromMatrix] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	capacities, err := c.orderMatrix.SelectCapacities(c.projectID, params.VenueTypeID)
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetCapacitiesFromMatrix] failed get all capacities, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get all capacitiess", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponseOrderMatrix, 0, len(capacities))
	for _, capacity := range capacities {
		res = append(res, view.DataResponseOrderMatrix{
			Type: "orderMatrix",
			Attributes: view.SummaryCapacityAttributes{
				Capacity: capacity.Capacity,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAgingsFromMatrix(w http.ResponseWriter, r *http.Request) {
	var params reqSumAgings
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handleGetAgingsFromMatrix] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	agings, err := c.orderMatrix.SelectAgings(c.projectID, params.VenueTypeID, params.Capacity)
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetAgingsFromMatrix] failed get all agings, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get all agings", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponseOrderMatrix, 0, len(agings))
	for _, aging := range agings {
		res = append(res, view.DataResponseOrderMatrix{
			Type: "orderMatrix",
			Attributes: view.SummaryAgingAttributes{
				AgingID:   aging.AgingID,
				AgingName: aging.AgingName,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetDevicesFromMatrix(w http.ResponseWriter, r *http.Request) {
	var params reqSumDevices
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handleGetDevicesFromMatrix] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	devices, err := c.orderMatrix.SelectDevices(c.projectID, params.VenueTypeID, params.Capacity, params.AgingID)
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetDevicesFromMatrix] failed get all Devices, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get all devices", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponseOrderMatrix, 0, len(devices))
	for _, device := range devices {
		res = append(res, view.DataResponseOrderMatrix{
			Type: "orderMatrix",
			Attributes: view.SummaryDeviceAttributes{
				DeviceID:   device.DeviceID,
				DeviceName: device.DeviceName,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}
