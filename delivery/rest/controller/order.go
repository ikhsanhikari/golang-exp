package controller

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	// "git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
)

func (c *Controller) handlePostOrder(w http.ResponseWriter, r *http.Request) {
	var (
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		params reqOrderInsert
	)

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePostOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	lastOrderNumber, err := c.order.GetLastOrderNumber()
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] failed get last order number, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get last order number", http.StatusInternalServerError)
		return
	}

	dateNow := time.Now().Format("060102")
	if strings.Compare(dateNow, lastOrderNumber.Date) == 1 {
		lastOrderNumber.Number = 0
	}
	orderNumber := "MN" + dateNow + leftPadLen(strconv.FormatInt((lastOrderNumber.Number+1), 10), "0", 7)

	totalPrice, err := c.calculateTotalPrice(params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity, params.AgingID, 10)
	if err != nil {
		c.reporter.Errorf("[handlePostOrder] failed calculate total price, err: %s", err.Error())
		view.RenderJSONError(w, "Failed calculate total price", http.StatusInternalServerError)
		return
	}

	order := order.Order{
		OrderNumber:     orderNumber,
		BuyerID:         123,
		VenueID:         params.VenueID,
		DeviceID:        params.DeviceID,
		ProductID:       params.ProductID,
		InstallationID:  params.InstallationID,
		Quantity:        params.Quantity,
		AgingID:         params.AgingID,
		RoomID:          params.RoomID,
		RoomQuantity:    params.RoomQuantity,
		TotalPrice:      totalPrice,
		PaymentMethodID: params.PaymentMethodID,
		PaymentFee:      params.PaymentFee,
		ProjectID:       10,
	}

	err = c.order.Insert(&order)
	if err != nil {
		c.reporter.Errorf("[handlePostOrder] failed post order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post order", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID:   order.OrderID,
		Type: "order",
		Attributes: view.OrderAttributesInsert{
			VenueID:         order.VenueID,
			DeviceID:        order.DeviceID,
			ProductID:       order.ProductID,
			InstallationID:  order.InstallationID,
			Quantity:        order.Quantity,
			AgingID:         order.AgingID,
			RoomID:          order.RoomID,
			RoomQuantity:    order.RoomQuantity,
			PaymentMethodID: order.PaymentMethodID,
			PaymentFee:      order.PaymentFee,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handlePatchOrder(w http.ResponseWriter, r *http.Request) {
	var (
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		params  reqOrderUpdate
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	getOrder, err := c.order.Get(id, 10)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Failed get order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	totalPrice, err := c.calculateTotalPrice(params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity, params.AgingID, 10)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] failed calculate total price, err: %s", err.Error())
		view.RenderJSONError(w, "Failed calculate total price", http.StatusInternalServerError)
		return
	}

	order := order.Order{
		OrderID:         id,
		VenueID:         params.VenueID,
		DeviceID:        params.DeviceID,
		ProductID:       params.ProductID,
		InstallationID:  params.InstallationID,
		Quantity:        params.Quantity,
		AgingID:         params.AgingID,
		RoomID:          params.RoomID,
		RoomQuantity:    params.RoomQuantity,
		TotalPrice:      totalPrice,
		PaymentMethodID: params.PaymentMethodID,
		PaymentFee:      params.PaymentFee,
		Status:          params.Status,
		ProjectID:       10,
		PendingAt:       getOrder.PendingAt,
		PaidAt:          getOrder.PaidAt,
		FailedAt:        getOrder.FailedAt,
	}

	err = c.order.Update(&order)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] failed update order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update order", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID:   order.OrderID,
		Type: "order",
		Attributes: view.OrderAttributesUpdate{
			VenueID:         order.VenueID,
			DeviceID:        order.DeviceID,
			ProductID:       order.ProductID,
			InstallationID:  order.InstallationID,
			Quantity:        order.Quantity,
			AgingID:         order.AgingID,
			RoomID:          order.RoomID,
			RoomQuantity:    order.RoomQuantity,
			PaymentMethodID: order.PaymentMethodID,
			PaymentFee:      order.PaymentFee,
			Status:          order.Status,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleUpdateStatusOrderByID(w http.ResponseWriter, r *http.Request) {
	var (
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		params  reqUpdateOrderStatus
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleUpdateStatusOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	getOrder, err := c.order.Get(id, 10)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleUpdateStatusOrder] order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleUpdateStatusOrder] Failed get order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handleUpdateStatusOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	order := order.Order{
		OrderID:   id,
		ProjectID: 10,
		Status:    params.Status,
		PendingAt: getOrder.PendingAt,
		PaidAt:    getOrder.PaidAt,
		FailedAt:  getOrder.FailedAt,
	}

	err = c.order.UpdateStatus(&order)
	if err != nil {
		c.reporter.Errorf("[handleUpdateStatusOrder] failed update status order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update order", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID:   order.OrderID,
		Type: "order",
		Attributes: view.OrderAttributesUpdateStatus{
			Status: order.Status,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteOrder(w http.ResponseWriter, r *http.Request) {
	var (
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleDeleteOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.order.Get(id, 10)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteOrder] order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteOrder] failed get order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
		return
	}

	err = c.order.Delete(id, 10)
	if err != nil {
		c.reporter.Errorf("[handleDeleteOrder] failed delete order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete order", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID: id,
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllOrders(w http.ResponseWriter, r *http.Request) {
	var (
	// project, _ = authpassport.GetProject(r)
	// pid        = project.ID
	)

	orders, err := c.order.Select(10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllOrders] orders not found, err: %s", err.Error())
		view.RenderJSONError(w, "Orders not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetAllOrders] failed get orders, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get orders", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponseOrder, 0, len(orders))
	for _, order := range orders {
		res = append(res, view.DataResponseOrder{
			ID:   order.OrderID,
			Type: "orders",
			Attributes: view.OrderAttributes{
				OrderNumber:     order.OrderNumber,
				BuyerID:         order.BuyerID,
				VenueID:         order.VenueID,
				DeviceID:        order.DeviceID,
				ProductID:       order.ProductID,
				InstallationID:  order.InstallationID,
				Quantity:        order.Quantity,
				AgingID:         order.AgingID,
				RoomID:          order.RoomID,
				RoomQuantity:    order.RoomQuantity,
				TotalPrice:      order.TotalPrice,
				PaymentMethodID: order.PaymentMethodID,
				PaymentFee:      order.PaymentFee,
				Status:          order.Status,
				CreatedAt:       order.CreatedAt,
				UpdatedAt:       order.UpdatedAt,
				DeletedAt:       order.DeletedAt,
				PendingAt:       order.PendingAt,
				PaidAt:          order.PaidAt,
				FailedAt:        order.FailedAt,
				ProjectID:       order.ProjectID,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	var (
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleGetOrderByID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	order, err := c.order.Get(id, 10)
	if err != nil {
		c.reporter.Errorf("[handleGetOrderByID] order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Orders not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetOrderByID] failed get order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get orders", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID:   order.OrderID,
		Type: "order",
		Attributes: view.OrderAttributes{
			OrderNumber:     order.OrderNumber,
			BuyerID:         order.BuyerID,
			VenueID:         order.VenueID,
			DeviceID:        order.DeviceID,
			ProductID:       order.ProductID,
			InstallationID:  order.InstallationID,
			Quantity:        order.Quantity,
			AgingID:         order.AgingID,
			RoomID:          order.RoomID,
			RoomQuantity:    order.RoomQuantity,
			TotalPrice:      order.TotalPrice,
			PaymentMethodID: order.PaymentMethodID,
			PaymentFee:      order.PaymentFee,
			Status:          order.Status,
			CreatedAt:       order.CreatedAt,
			UpdatedAt:       order.UpdatedAt,
			DeletedAt:       order.DeletedAt,
			PendingAt:       order.PendingAt,
			PaidAt:          order.PaidAt,
			FailedAt:        order.FailedAt,
			ProjectID:       order.ProjectID,
		},
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllByVenueID(w http.ResponseWriter, r *http.Request) {
	venue_id, err := strconv.ParseInt(router.GetParam(r, "venue_id"), 10, 64)

	orders, err := c.order.SelectByVenueID(venue_id, 10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllOrdersByVenueID] orders not found, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get orders", http.StatusInternalServerError)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetAllOrdersByVenueID] failed get orders, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get orders", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(orders))
	for _, order := range orders {
		res = append(res, view.DataResponse{
			Type: "orders",
			ID:   order.OrderID,
			Attributes: view.OrderAttributes{
				OrderNumber:     order.OrderNumber,
				BuyerID:         order.BuyerID,
				VenueID:         order.VenueID,
				DeviceID:        order.DeviceID,
				ProductID:       order.ProductID,
				InstallationID:  order.InstallationID,
				Quantity:        order.Quantity,
				AgingID:         order.AgingID,
				RoomID:          order.RoomID,
				RoomQuantity:    order.RoomQuantity,
				TotalPrice:      order.TotalPrice,
				PaymentMethodID: order.PaymentMethodID,
				PaymentFee:      order.PaymentFee,
				Status:          order.Status,
				CreatedAt:       order.CreatedAt,
				UpdatedAt:       order.UpdatedAt,
				DeletedAt:       order.DeletedAt,
				PendingAt:       order.PendingAt,
				PaidAt:          order.PaidAt,
				FailedAt:        order.FailedAt,
				ProjectID:       order.ProjectID,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllByBuyerID(w http.ResponseWriter, r *http.Request) {
	buyer_id, err := strconv.ParseInt(router.GetParam(r, "buyer_id"), 10, 64)

	orders, err := c.order.SelectByBuyerID(buyer_id, 10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllOrdersByBuyerID] orders not found, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get orders", http.StatusInternalServerError)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetAllOrdersByBuyerID] failed get orders, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get orders", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(orders))
	for _, order := range orders {
		res = append(res, view.DataResponse{
			Type: "orders",
			ID:   order.OrderID,
			Attributes: view.OrderAttributes{
				OrderNumber:     order.OrderNumber,
				BuyerID:         order.BuyerID,
				VenueID:         order.VenueID,
				DeviceID:        order.DeviceID,
				ProductID:       order.ProductID,
				InstallationID:  order.InstallationID,
				Quantity:        order.Quantity,
				AgingID:         order.AgingID,
				RoomID:          order.RoomID,
				RoomQuantity:    order.RoomQuantity,
				TotalPrice:      order.TotalPrice,
				PaymentMethodID: order.PaymentMethodID,
				PaymentFee:      order.PaymentFee,
				Status:          order.Status,
				CreatedAt:       order.CreatedAt,
				UpdatedAt:       order.UpdatedAt,
				DeletedAt:       order.DeletedAt,
				PendingAt:       order.PendingAt,
				PaidAt:          order.PaidAt,
				FailedAt:        order.FailedAt,
				ProjectID:       order.ProjectID,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllByPaidDate(w http.ResponseWriter, r *http.Request) {
	paiddate := router.GetParam(r, "paid_date")

	//layout := "2006-01-02T15:04:05"
	//t, err := time.Parse(layout, paiddate)
	paidd := paiddate[:10]
	orders, err := c.order.SelectByPaidDate(paidd, 10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllOrdersByPaidDate] orders not found, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get orders", http.StatusInternalServerError)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetAllOrdersByPaidDate] failed get order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(orders))
	for _, order := range orders {
		res = append(res, view.DataResponse{
			Type: "orders",
			ID:   order.OrderID,
			Attributes: view.OrderAttributes{
				OrderNumber:     order.OrderNumber,
				BuyerID:         order.BuyerID,
				VenueID:         order.VenueID,
				DeviceID:        order.DeviceID,
				ProductID:       order.ProductID,
				InstallationID:  order.InstallationID,
				Quantity:        order.Quantity,
				AgingID:         order.AgingID,
				RoomID:          order.RoomID,
				RoomQuantity:    order.RoomQuantity,
				TotalPrice:      order.TotalPrice,
				PaymentMethodID: order.PaymentMethodID,
				PaymentFee:      order.PaymentFee,
				Status:          order.Status,
				CreatedAt:       order.CreatedAt,
				UpdatedAt:       order.UpdatedAt,
				DeletedAt:       order.DeletedAt,
				PendingAt:       order.PendingAt,
				PaidAt:          order.PaidAt,
				FailedAt:        order.FailedAt,
				ProjectID:       order.ProjectID,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func leftPadLen(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

func (c *Controller) calculateTotalPrice(deviceID int64, productID int64, installationID int64, roomID int64, roomQuantity int64, agingID int64, pid int64) (float64, error) {
	device, err := c.device.Get(pid, deviceID)
	if err == sql.ErrNoRows {
		return 0, err
	}

	product, err := c.product.Get(pid, productID)
	if err == sql.ErrNoRows {
		return 0, err
	}

	installation, err := c.installation.Get(installationID, pid)
	if err == sql.ErrNoRows {
		return 0, err
	}

	room, err := c.room.Get(pid, roomID)
	if err == sql.ErrNoRows {
		return 0, err
	}

	aging, err := c.aging.Get(agingID, pid)
	if err == sql.ErrNoRows {
		return 0, err
	}

	return (device.Price + product.Price + installation.Price + (room.Price * float64(roomQuantity)) + aging.Price), err
}
