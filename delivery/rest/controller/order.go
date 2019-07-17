package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
)

func (c *Controller) handlePostOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostOrder] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handlePostOrder] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	var params reqOrder
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePostOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.venue.Get(10, params.VenueID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Venue Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue Not Found", http.StatusNotFound)
		return
	}

	device, err := c.device.Get(10, params.DeviceID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Device Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Device Not Found", http.StatusNotFound)
		return
	}

	product, err := c.product.Get(10, params.ProductID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Product Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Product Not Found", http.StatusNotFound)
		return
	}

	installation, err := c.installation.Get(params.InstallationID, 10)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Installation Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation Not Found", http.StatusNotFound)
		return
	}

	room, err := c.room.Get(10, params.RoomID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Room Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Room Not Found", http.StatusNotFound)
		return
	}

	aging, err := c.aging.Get(params.AgingID, 10)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Aging Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging Not Found", http.StatusNotFound)
		return
	}

	//generate order number
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

	//calculate total price
	totalPrice := c.calculateTotalPrice(device.Price, product.Price, installation.Price, room.Price, params.RoomQuantity, aging.Price)

	//insert order
	insertOrder := order.Order{
		OrderNumber:    orderNumber,
		BuyerID:        fmt.Sprintf("%v", userID),
		VenueID:        params.VenueID,
		DeviceID:       params.DeviceID,
		ProductID:      params.ProductID,
		InstallationID: params.InstallationID,
		Quantity:       params.Quantity,
		AgingID:        params.AgingID,
		RoomID:         params.RoomID,
		RoomQuantity:   params.RoomQuantity,
		TotalPrice:     totalPrice,
		PaymentFee:     params.PaymentFee,
		CreatedBy:      fmt.Sprintf("%v", userID),
		LastUpdateBy:   fmt.Sprintf("%v", userID),
		ProjectID:      10,
		Email:          params.Email,
	}

	err = c.order.Insert(&insertOrder)
	if err != nil {
		c.reporter.Errorf("[handlePostOrder] failed post order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post order", http.StatusInternalServerError)
		return
	}

	//insert order details
	err = c.insertOrderDetail(insertOrder, device, product, installation, room, aging)
	if err != nil {
		c.reporter.Errorf("[handlePostOrder] failed post order details, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post order details", http.StatusInternalServerError)
		return
	}

	//set response
	res := view.DataResponseOrder{
		ID:   insertOrder.OrderID,
		Type: "order",
		Attributes: view.OrderAttributes{
			OrderNumber:       insertOrder.OrderNumber,
			BuyerID:           insertOrder.BuyerID,
			VenueID:           insertOrder.VenueID,
			DeviceID:          insertOrder.DeviceID,
			ProductID:         insertOrder.ProductID,
			InstallationID:    insertOrder.InstallationID,
			Quantity:          insertOrder.Quantity,
			AgingID:           insertOrder.AgingID,
			RoomID:            insertOrder.RoomID,
			RoomQuantity:      insertOrder.RoomQuantity,
			TotalPrice:        insertOrder.TotalPrice,
			PaymentMethodID:   insertOrder.PaymentMethodID,
			PaymentFee:        insertOrder.PaymentFee,
			Status:            insertOrder.Status,
			CreatedAt:         insertOrder.CreatedAt,
			CreatedBy:         insertOrder.CreatedBy,
			UpdatedAt:         insertOrder.UpdatedAt,
			LastUpdateBy:      insertOrder.LastUpdateBy,
			DeletedAt:         insertOrder.DeletedAt,
			PendingAt:         insertOrder.PendingAt,
			PaidAt:            insertOrder.PaidAt,
			FailedAt:          insertOrder.FailedAt,
			ProjectID:         insertOrder.ProjectID,
			Email:             insertOrder.Email,
			OpenPaymentStatus: insertOrder.OpenPaymentStatus,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handlePatchOrderForPayment(w http.ResponseWriter, r *http.Request) {
	var (
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrderForPayment] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePatchOrderForPayment] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handlePatchOrderForPayment] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	//check open payment status
	getOrder, err := c.order.Get(id, 10, fmt.Sprintf("%v", userID))
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrderForPayment] order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrderForPayment] Failed get order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
		return
	}
	if getOrder.OpenPaymentStatus == 0 {
		c.reporter.Errorf("[handlePatchOrderForPayment] Not Approved")
		view.RenderJSONError(w, "Not Approved", http.StatusBadRequest)
		return
	}

	//do payment
	payment, err := c.payment.Pay(strconv.FormatInt(getOrder.OrderID, 10), getOrder.PaymentMethodID)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrderForPayment] Failed processing payment, err: %s", err.Error())
		view.RenderJSONError(w, "Failed processing payment", http.StatusInternalServerError)
		return
	}
	if payment == nil {
		c.reporter.Errorf("[handlePatchOrderForPayment] Failed processing payment")
		view.RenderJSONError(w, "Failed processing payment", http.StatusInternalServerError)
		return
	}
	if payment.PaymentData.URL == "" {
		c.reporter.Errorf("[handlePatchOrderForPayment] Failed processing payment, URL not found")
		view.RenderJSONError(w, "Failed processing payment, URL not found", http.StatusInternalServerError)
		return
	}

	//update status = 1
	updateStatus := order.Order{
		OrderID:      getOrder.OrderID,
		ProjectID:    getOrder.ProjectID,
		Status:       1,
		CreatedBy:    getOrder.CreatedBy,
		LastUpdateBy: fmt.Sprintf("%v", userID),
		PendingAt:    getOrder.PendingAt,
		PaidAt:       getOrder.PaidAt,
		FailedAt:     getOrder.FailedAt,
		VenueID:      getOrder.VenueID,
		BuyerID:      getOrder.BuyerID,
	}

	err = c.order.UpdateOrderStatus(&updateStatus)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrderForPayment] failed update status order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update order", http.StatusInternalServerError)
		return
	}

	//set response
	res := view.DataResponseOrderPayment{
		ID:   updateStatus.OrderID,
		Type: "order",
		Attributes: view.OrderAttributes{
			OrderNumber:       getOrder.OrderNumber,
			BuyerID:           getOrder.BuyerID,
			VenueID:           getOrder.VenueID,
			DeviceID:          getOrder.DeviceID,
			ProductID:         getOrder.ProductID,
			InstallationID:    getOrder.InstallationID,
			Quantity:          getOrder.Quantity,
			AgingID:           getOrder.AgingID,
			RoomID:            getOrder.RoomID,
			RoomQuantity:      getOrder.RoomQuantity,
			TotalPrice:        getOrder.TotalPrice,
			PaymentMethodID:   getOrder.PaymentMethodID,
			PaymentFee:        getOrder.PaymentFee,
			Status:            updateStatus.Status,
			CreatedAt:         getOrder.CreatedAt,
			CreatedBy:         getOrder.CreatedBy,
			UpdatedAt:         updateStatus.UpdatedAt,
			LastUpdateBy:      updateStatus.LastUpdateBy,
			DeletedAt:         getOrder.DeletedAt,
			PendingAt:         updateStatus.PendingAt,
			PaidAt:            updateStatus.PaidAt,
			FailedAt:          updateStatus.FailedAt,
			ProjectID:         updateStatus.ProjectID,
			Email:             getOrder.Email,
			OpenPaymentStatus: getOrder.OpenPaymentStatus,
		},
		ResponseType:    payment.ResponseType,
		HTMLRedirection: payment.HTMLRedirection,
		PaymentData: view.PaymentAttributes{
			URL: payment.PaymentData.URL,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handlePatchOrder(w http.ResponseWriter, r *http.Request) {
	var (
		params  reqOrder
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePatchOrder] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handlePatchOrder] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	//validasi order
	getOrder, err := c.order.Get(id, 10, fmt.Sprintf("%v", userID))
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

	//validasi order detail
	orderDetails, err := c.orderDetail.Get(id, 10, userID.(string))
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] order details not found, err: %s", err.Error())
		view.RenderJSONError(w, "Order details not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Failed get order details, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order detail", http.StatusInternalServerError)
		return
	}
	if len(orderDetails) != 5 {
		c.reporter.Errorf("[handlePatchOrder] Failed get all order details")
		view.RenderJSONError(w, "Failed get all order details", http.StatusInternalServerError)
		return
	}

	//validasi request body
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	//validasi foreign key
	_, err = c.venue.Get(10, params.VenueID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Venue Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue Not Found", http.StatusNotFound)
		return
	}

	device, err := c.device.Get(10, params.DeviceID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Device Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Device Not Found", http.StatusNotFound)
		return
	}

	product, err := c.product.Get(10, params.ProductID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Product Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Product Not Found", http.StatusNotFound)
		return
	}

	installation, err := c.installation.Get(params.InstallationID, 10)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Installation Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation Not Found", http.StatusNotFound)
		return
	}

	room, err := c.room.Get(10, params.RoomID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Room Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Room Not Found", http.StatusNotFound)
		return
	}

	aging, err := c.aging.Get(params.AgingID, 10)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Aging Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging Not Found", http.StatusNotFound)
		return
	}

	//calculate total price
	totalPrice := c.calculateTotalPrice(device.Price, product.Price, installation.Price, room.Price, params.RoomQuantity, aging.Price)

	//update order
	updateOrder := order.Order{
		OrderID:        id,
		VenueID:        params.VenueID,
		DeviceID:       params.DeviceID,
		ProductID:      params.ProductID,
		InstallationID: params.InstallationID,
		Quantity:       params.Quantity,
		AgingID:        params.AgingID,
		RoomID:         params.RoomID,
		RoomQuantity:   params.RoomQuantity,
		TotalPrice:     totalPrice,
		PaymentFee:     params.PaymentFee,
		ProjectID:      10,
		Status:         getOrder.Status,
		CreatedBy:      getOrder.CreatedBy,
		LastUpdateBy:   fmt.Sprintf("%v", userID),
		Email:          params.Email,
	}

	err = c.order.Update(&updateOrder)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] failed update order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update order", http.StatusInternalServerError)
		return
	}

	//update order details
	err = c.updateOrderDetail(updateOrder, device, product, installation, room, aging)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] failed update order details, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update order details", http.StatusInternalServerError)
		return
	}

	//set response
	res := view.DataResponseOrder{
		ID:   updateOrder.OrderID,
		Type: "order",
		Attributes: view.OrderAttributes{
			OrderNumber:       getOrder.OrderNumber,
			BuyerID:           getOrder.BuyerID,
			VenueID:           updateOrder.VenueID,
			DeviceID:          updateOrder.DeviceID,
			ProductID:         updateOrder.ProductID,
			InstallationID:    updateOrder.InstallationID,
			Quantity:          updateOrder.Quantity,
			AgingID:           updateOrder.AgingID,
			RoomID:            updateOrder.RoomID,
			RoomQuantity:      updateOrder.RoomQuantity,
			TotalPrice:        updateOrder.TotalPrice,
			PaymentMethodID:   updateOrder.PaymentMethodID,
			PaymentFee:        updateOrder.PaymentFee,
			Status:            getOrder.Status,
			CreatedAt:         getOrder.CreatedAt,
			CreatedBy:         getOrder.CreatedBy,
			UpdatedAt:         updateOrder.UpdatedAt,
			LastUpdateBy:      updateOrder.LastUpdateBy,
			DeletedAt:         getOrder.DeletedAt,
			PendingAt:         getOrder.PendingAt,
			PaidAt:            getOrder.PaidAt,
			FailedAt:          getOrder.FailedAt,
			ProjectID:         updateOrder.ProjectID,
			Email:             updateOrder.Email,
			OpenPaymentStatus: getOrder.OpenPaymentStatus,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleUpdateOrderStatusByID(w http.ResponseWriter, r *http.Request) {
	var (
		params  reqUpdateOrderStatus
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleUpdateOrderStatus] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleUpdateStatusOrder] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		userID = ""
	}

	getOrder, err := c.order.Get(id, 10, fmt.Sprintf("%v", userID))
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleUpdateOrderStatus] order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleUpdateOrderStatus] Failed get order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handleUpdateOrderStatus] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	//update status order
	updateStatus := order.Order{
		OrderID:      id,
		ProjectID:    10,
		Status:       params.Status,
		CreatedBy:    getOrder.CreatedBy,
		LastUpdateBy: fmt.Sprintf("%v", userID),
		PendingAt:    getOrder.PendingAt,
		PaidAt:       getOrder.PaidAt,
		FailedAt:     getOrder.FailedAt,
		VenueID:      getOrder.VenueID,
		BuyerID:      getOrder.BuyerID,
	}

	err = c.order.UpdateOrderStatus(&updateStatus)
	if err != nil {
		c.reporter.Errorf("[handleUpdateOrderStatus] failed update order status, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update order status", http.StatusInternalServerError)
		return
	}

	//set response
	res := view.DataResponseOrder{
		ID:   updateStatus.OrderID,
		Type: "order",
		Attributes: view.OrderAttributes{
			OrderNumber:       getOrder.OrderNumber,
			BuyerID:           updateStatus.BuyerID,
			VenueID:           updateStatus.VenueID,
			DeviceID:          getOrder.DeviceID,
			ProductID:         getOrder.ProductID,
			InstallationID:    getOrder.InstallationID,
			Quantity:          getOrder.Quantity,
			AgingID:           getOrder.AgingID,
			RoomID:            getOrder.RoomID,
			RoomQuantity:      getOrder.RoomQuantity,
			TotalPrice:        getOrder.TotalPrice,
			PaymentMethodID:   getOrder.PaymentMethodID,
			PaymentFee:        getOrder.PaymentFee,
			Status:            updateStatus.Status,
			CreatedAt:         getOrder.CreatedAt,
			CreatedBy:         updateStatus.CreatedBy,
			UpdatedAt:         updateStatus.UpdatedAt,
			LastUpdateBy:      updateStatus.LastUpdateBy,
			DeletedAt:         getOrder.DeletedAt,
			PendingAt:         updateStatus.PendingAt,
			PaidAt:            updateStatus.PaidAt,
			FailedAt:          updateStatus.FailedAt,
			ProjectID:         updateStatus.ProjectID,
			Email:             getOrder.Email,
			OpenPaymentStatus: getOrder.OpenPaymentStatus,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleUpdateOpenPaymentStatusByID(w http.ResponseWriter, r *http.Request) {
	var (
		params  reqUpdateOpenPaymentStatus
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleUpdateOpenPaymentStatus] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handleUpdateOpenPaymentStatus] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleUpdateOpenPaymentStatus] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		if params.UserID == "" {
			c.reporter.Errorf("[handleUpdateOpenPaymentStatus] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
	}

	getOrder, err := c.order.Get(id, 10, fmt.Sprintf("%v", userID))
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleUpdateOpenPaymentStatus] order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleUpdateOpenPaymentStatus] Failed get order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
		return
	}

	updateStatus := order.Order{
		OrderID:           id,
		ProjectID:         10,
		OpenPaymentStatus: params.OpenPaymentStatus,
		CreatedBy:         getOrder.CreatedBy,
		LastUpdateBy:      fmt.Sprintf("%v", userID),
		VenueID:           getOrder.VenueID,
		BuyerID:           getOrder.BuyerID,
	}

	err = c.order.UpdateOpenPaymentStatus(&updateStatus)
	if err != nil {
		c.reporter.Errorf("[handleUpdateOpenPaymentStatus] failed update open payment status, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update open payment status", http.StatusInternalServerError)
		return
	}

	res := view.DataResponseOrder{
		ID:   updateStatus.OrderID,
		Type: "order",
		Attributes: view.OrderAttributes{
			OrderNumber:       getOrder.OrderNumber,
			BuyerID:           updateStatus.BuyerID,
			VenueID:           updateStatus.VenueID,
			DeviceID:          getOrder.DeviceID,
			ProductID:         getOrder.ProductID,
			InstallationID:    getOrder.InstallationID,
			Quantity:          getOrder.Quantity,
			AgingID:           getOrder.AgingID,
			RoomID:            getOrder.RoomID,
			RoomQuantity:      getOrder.RoomQuantity,
			TotalPrice:        getOrder.TotalPrice,
			PaymentMethodID:   getOrder.PaymentMethodID,
			PaymentFee:        getOrder.PaymentFee,
			Status:            getOrder.Status,
			CreatedAt:         getOrder.CreatedAt,
			CreatedBy:         updateStatus.CreatedBy,
			UpdatedAt:         updateStatus.UpdatedAt,
			LastUpdateBy:      updateStatus.LastUpdateBy,
			DeletedAt:         getOrder.DeletedAt,
			PendingAt:         getOrder.PendingAt,
			PaidAt:            getOrder.PaidAt,
			FailedAt:          getOrder.FailedAt,
			ProjectID:         updateStatus.ProjectID,
			Email:             getOrder.Email,
			OpenPaymentStatus: updateStatus.OpenPaymentStatus,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteOrder(w http.ResponseWriter, r *http.Request) {
	var (
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleDeleteOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteOrder] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleDeleteOrder] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	//validasi order
	getOrder, err := c.order.Get(id, 10, fmt.Sprintf("%v", userID))
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

	//validasi order detail
	orderDetails, err := c.orderDetail.Get(id, 10, userID.(string))
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] order details not found, err: %s", err.Error())
		view.RenderJSONError(w, "Order details not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Failed get order details, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order detail", http.StatusInternalServerError)
		return
	}
	if len(orderDetails) != 5 {
		c.reporter.Errorf("[handlePatchOrder] Failed get all order details")
		view.RenderJSONError(w, "Failed get all order details", http.StatusInternalServerError)
		return
	}

	//delete order
	deleteOrder := order.Order{
		OrderID:      id,
		ProjectID:    10,
		CreatedBy:    getOrder.CreatedBy,
		LastUpdateBy: fmt.Sprintf("%v", userID),
		BuyerID:      getOrder.BuyerID,
		VenueID:      getOrder.VenueID,
		Status:       getOrder.Status,
		PaidAt:       getOrder.PaidAt,
	}

	err = c.order.Delete(&deleteOrder)
	if err != nil {
		c.reporter.Errorf("[handleDeleteOrder] failed delete order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete order", http.StatusInternalServerError)
		return
	}

	//delete order details
	err = c.deleteOrderDetail(deleteOrder)
	if err != nil {
		c.reporter.Errorf("[handleDeleteOrder] failed delete order details, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete order details", http.StatusInternalServerError)
		return
	}

	res := view.DataResponseOrder{
		ID: id,
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetAllOrder] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleGetAllOrder] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	orders, err := c.order.Select(10, fmt.Sprintf("%v", userID))
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
			Type: "order",
			Attributes: view.OrderAttributes{
				OrderNumber:       order.OrderNumber,
				BuyerID:           order.BuyerID,
				VenueID:           order.VenueID,
				DeviceID:          order.DeviceID,
				ProductID:         order.ProductID,
				InstallationID:    order.InstallationID,
				Quantity:          order.Quantity,
				AgingID:           order.AgingID,
				RoomID:            order.RoomID,
				RoomQuantity:      order.RoomQuantity,
				TotalPrice:        order.TotalPrice,
				PaymentMethodID:   order.PaymentMethodID,
				PaymentFee:        order.PaymentFee,
				Status:            order.Status,
				CreatedAt:         order.CreatedAt,
				CreatedBy:         order.CreatedBy,
				UpdatedAt:         order.UpdatedAt,
				LastUpdateBy:      order.LastUpdateBy,
				DeletedAt:         order.DeletedAt,
				PendingAt:         order.PendingAt,
				PaidAt:            order.PaidAt,
				FailedAt:          order.FailedAt,
				ProjectID:         order.ProjectID,
				Email:             order.Email,
				OpenPaymentStatus: order.OpenPaymentStatus,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	var (
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleGetOrderByID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetOrderByID] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		userID = ""
	}

	order, err := c.order.Get(id, 10, fmt.Sprintf("%v", userID))
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

	res := view.DataResponseOrder{
		ID:   order.OrderID,
		Type: "order",
		Attributes: view.OrderAttributes{
			OrderNumber:       order.OrderNumber,
			BuyerID:           order.BuyerID,
			VenueID:           order.VenueID,
			DeviceID:          order.DeviceID,
			ProductID:         order.ProductID,
			InstallationID:    order.InstallationID,
			Quantity:          order.Quantity,
			AgingID:           order.AgingID,
			RoomID:            order.RoomID,
			RoomQuantity:      order.RoomQuantity,
			TotalPrice:        order.TotalPrice,
			PaymentMethodID:   order.PaymentMethodID,
			PaymentFee:        order.PaymentFee,
			Status:            order.Status,
			CreatedAt:         order.CreatedAt,
			CreatedBy:         order.CreatedBy,
			UpdatedAt:         order.UpdatedAt,
			LastUpdateBy:      order.LastUpdateBy,
			DeletedAt:         order.DeletedAt,
			PendingAt:         order.PendingAt,
			PaidAt:            order.PaidAt,
			FailedAt:          order.FailedAt,
			ProjectID:         order.ProjectID,
			Email:             order.Email,
			OpenPaymentStatus: order.OpenPaymentStatus,
		},
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetSumOrderByID(w http.ResponseWriter, r *http.Request) {
	var (
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleGetSumOrderByID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetSumOrderByID] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		userID = ""
		// c.reporter.Errorf("[handleGetOrderByID] failed get userID")
		// view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		// return
	}

	sumorder, err := c.order.SelectSummaryOrderByID(id, 10, fmt.Sprintf("%v", userID))
	if err != nil {
		c.reporter.Errorf("[handleGetSumOrderByID] sum order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Sum Order not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetSumOrderByID] failed get sum order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get sum order", http.StatusInternalServerError)
		return
	}

	res := view.DataResponseOrder{
		ID:   sumorder.OrderID,
		Type: "sum_order",
		Attributes: view.SumOrderAttributes{
			OrderNumber:        sumorder.OrderNumber,
			OrderTotalPrice:    sumorder.OrderTotalPrice,
			OrderCreatedAt:     sumorder.OrderCreatedAt,
			OrderPaidAt:        sumorder.OrderPaidAt,
			OrderFailedAt:      sumorder.OrderFailedAt,
			OrderEmail:         sumorder.OrderEmail,
			VenueName:          sumorder.VenueName,
			VenueType:          sumorder.VenueType,
			VenueAddress:       sumorder.VenueAddress,
			VenueProvince:      sumorder.VenueProvince,
			VenueZip:           sumorder.VenueZip,
			VenueCapacity:      sumorder.VenueCapacity,
			VenueLongitude:     sumorder.VenueLongitude,
			VenueLatitude:      sumorder.VenueLatitude,
			VenueCategory:      sumorder.VenueCategory,
			DeviceName:         sumorder.DeviceName,
			ProductName:        sumorder.ProductName,
			InstallationName:   sumorder.InstallationName,
			RoomName:           sumorder.RoomName,
			RoomQty:            sumorder.RoomQty,
			AgingName:          sumorder.AgingName,
			OrderStatus:        sumorder.OrderStatus,
			OpenPaymentStatus:  sumorder.OpenPaymentStatus,
			LicenseNumber:      sumorder.LicenseNumber,
			LicenseActiveDate:  sumorder.LicenseActiveDate,
			LicenseExpiredDate: sumorder.LicenseExpiredDate,
		},
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetSumOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetSumOrdersByUserID] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleGetSumOrdersByUserID] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	sumorders, err := c.order.SelectSummaryOrdersByUserID(10, fmt.Sprintf("%v", userID))
	if err != nil {
		c.reporter.Errorf("[handleGetSumOrdersByUserID] order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Sum orders not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetSumOrdersByUserID] failed get sum order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get sum orders", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponseOrder, 0, len(sumorders))
	for _, sumorder := range sumorders {
		res = append(res, view.DataResponseOrder{
			ID:   sumorder.OrderID,
			Type: "sum_order",
			Attributes: view.SumOrderAttributes{
				OrderNumber:        sumorder.OrderNumber,
				OrderTotalPrice:    sumorder.OrderTotalPrice,
				OrderCreatedAt:     sumorder.OrderCreatedAt,
				OrderPaidAt:        sumorder.OrderPaidAt,
				OrderFailedAt:      sumorder.OrderFailedAt,
				OrderEmail:         sumorder.OrderEmail,
				VenueName:          sumorder.VenueName,
				VenueType:          sumorder.VenueType,
				VenueAddress:       sumorder.VenueAddress,
				VenueProvince:      sumorder.VenueProvince,
				VenueZip:           sumorder.VenueZip,
				VenueCapacity:      sumorder.VenueCapacity,
				VenueLongitude:     sumorder.VenueLongitude,
				VenueLatitude:      sumorder.VenueLatitude,
				VenueCategory:      sumorder.VenueCategory,
				DeviceName:         sumorder.DeviceName,
				ProductName:        sumorder.ProductName,
				InstallationName:   sumorder.InstallationName,
				RoomName:           sumorder.RoomName,
				RoomQty:            sumorder.RoomQty,
				AgingName:          sumorder.AgingName,
				OrderStatus:        sumorder.OrderStatus,
				OpenPaymentStatus:  sumorder.OpenPaymentStatus,
				LicenseNumber:      sumorder.LicenseNumber,
				LicenseActiveDate:  sumorder.LicenseActiveDate,
				LicenseExpiredDate: sumorder.LicenseExpiredDate,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllByVenueID(w http.ResponseWriter, r *http.Request) {
	var (
		_venueID     = router.GetParam(r, "venue_id")
		venueID, err = strconv.ParseInt(_venueID, 10, 64)
	)
	if err != nil {
		c.reporter.Errorf("[handleGetAllOrdersByVenueID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetAllOrdersByVenueID] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleGetAllOrdersByVenueID] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	orders, err := c.order.SelectByVenueID(venueID, 10, fmt.Sprintf("%v", userID))
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

	res := make([]view.DataResponseOrder, 0, len(orders))
	for _, order := range orders {
		res = append(res, view.DataResponseOrder{
			Type: "order",
			ID:   order.OrderID,
			Attributes: view.OrderAttributes{
				OrderNumber:       order.OrderNumber,
				BuyerID:           order.BuyerID,
				VenueID:           order.VenueID,
				DeviceID:          order.DeviceID,
				ProductID:         order.ProductID,
				InstallationID:    order.InstallationID,
				Quantity:          order.Quantity,
				AgingID:           order.AgingID,
				RoomID:            order.RoomID,
				RoomQuantity:      order.RoomQuantity,
				TotalPrice:        order.TotalPrice,
				PaymentMethodID:   order.PaymentMethodID,
				PaymentFee:        order.PaymentFee,
				Status:            order.Status,
				CreatedAt:         order.CreatedAt,
				CreatedBy:         order.CreatedBy,
				UpdatedAt:         order.UpdatedAt,
				LastUpdateBy:      order.LastUpdateBy,
				DeletedAt:         order.DeletedAt,
				PendingAt:         order.PendingAt,
				PaidAt:            order.PaidAt,
				FailedAt:          order.FailedAt,
				ProjectID:         order.ProjectID,
				Email:             order.Email,
				OpenPaymentStatus: order.OpenPaymentStatus,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllByBuyerID(w http.ResponseWriter, r *http.Request) {
	var (
		buyerID = router.GetParam(r, "buyer_id")
	)

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetAllOrdersByBuyerID] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleGetAllOrdersByBuyerID] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	orders, err := c.order.SelectByBuyerID(buyerID, 10, fmt.Sprintf("%v", userID))
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

	res := make([]view.DataResponseOrder, 0, len(orders))
	for _, order := range orders {
		res = append(res, view.DataResponseOrder{
			Type: "order",
			ID:   order.OrderID,
			Attributes: view.OrderAttributes{
				OrderNumber:       order.OrderNumber,
				BuyerID:           order.BuyerID,
				VenueID:           order.VenueID,
				DeviceID:          order.DeviceID,
				ProductID:         order.ProductID,
				InstallationID:    order.InstallationID,
				Quantity:          order.Quantity,
				AgingID:           order.AgingID,
				RoomID:            order.RoomID,
				RoomQuantity:      order.RoomQuantity,
				TotalPrice:        order.TotalPrice,
				PaymentMethodID:   order.PaymentMethodID,
				PaymentFee:        order.PaymentFee,
				Status:            order.Status,
				CreatedAt:         order.CreatedAt,
				CreatedBy:         order.CreatedBy,
				UpdatedAt:         order.UpdatedAt,
				LastUpdateBy:      order.LastUpdateBy,
				DeletedAt:         order.DeletedAt,
				PendingAt:         order.PendingAt,
				PaidAt:            order.PaidAt,
				FailedAt:          order.FailedAt,
				ProjectID:         order.ProjectID,
				Email:             order.Email,
				OpenPaymentStatus: order.OpenPaymentStatus,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllByPaidDate(w http.ResponseWriter, r *http.Request) {
	var (
		_paidDate = router.GetParam(r, "paid_date")
		paidDate  = _paidDate[:10]
	)

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetAllOrdersByPaidDate] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleGetAllOrdersByPaidDate] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	orders, err := c.order.SelectByPaidDate(paidDate, 10, fmt.Sprintf("%v", userID))
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

	res := make([]view.DataResponseOrder, 0, len(orders))
	for _, order := range orders {
		res = append(res, view.DataResponseOrder{
			Type: "order",
			ID:   order.OrderID,
			Attributes: view.OrderAttributes{
				OrderNumber:       order.OrderNumber,
				BuyerID:           order.BuyerID,
				VenueID:           order.VenueID,
				DeviceID:          order.DeviceID,
				ProductID:         order.ProductID,
				InstallationID:    order.InstallationID,
				Quantity:          order.Quantity,
				AgingID:           order.AgingID,
				RoomID:            order.RoomID,
				RoomQuantity:      order.RoomQuantity,
				TotalPrice:        order.TotalPrice,
				PaymentMethodID:   order.PaymentMethodID,
				PaymentFee:        order.PaymentFee,
				Status:            order.Status,
				CreatedAt:         order.CreatedAt,
				CreatedBy:         order.CreatedBy,
				UpdatedAt:         order.UpdatedAt,
				LastUpdateBy:      order.LastUpdateBy,
				DeletedAt:         order.DeletedAt,
				PendingAt:         order.PendingAt,
				PaidAt:            order.PaidAt,
				FailedAt:          order.FailedAt,
				ProjectID:         order.ProjectID,
				Email:             order.Email,
				OpenPaymentStatus: order.OpenPaymentStatus,
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

func (c *Controller) calculateTotalPrice(devicePrice float64, productPrice float64, installationPrice float64, roomPrice float64, roomQuantity int64, agingPrice float64) float64 {

	return (devicePrice + productPrice + installationPrice + (roomPrice * float64(roomQuantity)) + agingPrice)
}
