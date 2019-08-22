package controller

import (
	"database/sql"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/room"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handlePostOrder(w http.ResponseWriter, r *http.Request) {
	var (
		params    reqOrder
		projectID = int64(10)
		isAdmin   = false
	)

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePostOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostOrder] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		if params.UserID == "" {
			c.reporter.Errorf("[handlePostOrder] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	var venue venue.Venue
	if isAdmin {
		venue, err = c.venue.Get(projectID, params.VenueID, "")
	} else {
		venue, err = c.venue.Get(projectID, params.VenueID, userID.(string))
	}
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Venue Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue Not Found", http.StatusNotFound)
		return
	}

	device, err := c.device.Get(projectID, params.DeviceID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Device Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Device Not Found", http.StatusNotFound)
		return
	}

	product, err := c.product.Get(projectID, params.ProductID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Product Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Product Not Found", http.StatusNotFound)
		return
	}

	installation, err := c.installation.Get(params.InstallationID, projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Installation Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation Not Found", http.StatusNotFound)
		return
	}

	var room room.Room
	if params.RoomID != 0 && params.RoomQuantity != 0 {
		room, err = c.room.Get(projectID, params.RoomID)
		if err == sql.ErrNoRows {
			c.reporter.Errorf("[handlePostOrder] Room Not Found, err: %s", err.Error())
			view.RenderJSONError(w, "Room Not Found", http.StatusNotFound)
			return
		}
	}

	aging, err := c.aging.Get(params.AgingID, projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Aging Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging Not Found", http.StatusNotFound)
		return
	}

	valid := isOrderValid(venue.VenueType, venue.Capacity, params.AgingID, params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity)
	if !valid {
		c.reporter.Errorf("[handlePostOrder] Order not valid, venueType: %d, capacity: %d, agingID: %d, deviceID: %d, productID: %d, installationID: %d, roomID: %d, roomQuantity: %d",
			venue.VenueType, venue.Capacity, params.AgingID, params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity)
		view.RenderJSONError(w, "Order not valid", http.StatusBadRequest)
		return
	}

	//generate order number
	orderNumber, err := c.generateOrderNumber()
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrder] Failed generate order number, err: %s", err.Error())
		view.RenderJSONError(w, "Failed generate order number", http.StatusInternalServerError)
		return
	}

	//calculate total price
	totalPrice := c.calculateTotalPrice(venue.VenueType, product.Price, installation.Price, room.Price, float64(params.RoomQuantity))

	//insert order
	insertOrder := order.Order{
		OrderNumber:    orderNumber,
		BuyerID:        userID.(string),
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
		Status:         0,
		CreatedBy:      userID.(string),
		LastUpdateBy:   userID.(string),
		ProjectID:      projectID,
		Email:          params.Email,
	}

	err = c.order.Insert(&insertOrder, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePostOrder] failed post order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post order", http.StatusInternalServerError)
		return
	}

	//insert order details
	err = c.insertOrderDetail(insertOrder, device, product, installation, room, aging, isAdmin)
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

func (c *Controller) handlePostOrderByAgent(w http.ResponseWriter, r *http.Request) {
	var (
		params    reqOrder
		projectID = int64(10)
		isAdmin   = false
	)

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handlePostOrderByAgent] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostOrderByAgent] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		if params.UserID == "" {
			c.reporter.Errorf("[handlePostOrderByAgent] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	//check admin
	if !isAdmin {
		_, isExist := c.admin.Check(userID.(string))
		if isExist == sql.ErrNoRows {
			c.reporter.Errorf("[handlePostOrderByAgent] user is not exist")
			view.RenderJSONError(w, "user is not exist", http.StatusUnauthorized)
			return
		}
	}

	var venue venue.Venue
	if isAdmin {
		venue, err = c.venue.Get(projectID, params.VenueID, "")
	} else {
		venue, err = c.venue.Get(projectID, params.VenueID, userID.(string))
	}
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderByAgent] Venue Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue Not Found", http.StatusNotFound)
		return
	}

	device, err := c.device.Get(projectID, params.DeviceID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderByAgent] Device Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Device Not Found", http.StatusNotFound)
		return
	}

	product, err := c.product.Get(projectID, params.ProductID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderByAgent] Product Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Product Not Found", http.StatusNotFound)
		return
	}

	installation, err := c.installation.Get(params.InstallationID, projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderByAgent] Installation Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation Not Found", http.StatusNotFound)
		return
	}

	var room room.Room
	if params.RoomID != 0 && params.RoomQuantity != 0 {
		room, err = c.room.Get(projectID, params.RoomID)
		if err == sql.ErrNoRows {
			c.reporter.Errorf("[handlePostOrderByAgent] Room Not Found, err: %s", err.Error())
			view.RenderJSONError(w, "Room Not Found", http.StatusNotFound)
			return
		}
	}

	aging, err := c.aging.Get(params.AgingID, projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderByAgent] Aging Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging Not Found", http.StatusNotFound)
		return
	}

	valid := isOrderValid(venue.VenueType, venue.Capacity, params.AgingID, params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity)
	if !valid {
		c.reporter.Errorf("[handlePostOrderByAgent] Order not valid, venueType: %d, capacity: %d, agingID: %d, deviceID: %d, productID: %d, installationID: %d, roomID: %d, roomQuantity: %d",
			venue.VenueType, venue.Capacity, params.AgingID, params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity)
		view.RenderJSONError(w, "Order not valid", http.StatusBadRequest)
		return
	}

	orderNumber, err := c.generateOrderNumber()
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePostOrderByAgent] Failed generate order number, err: %s", err.Error())
		view.RenderJSONError(w, "Failed generate order number", http.StatusInternalServerError)
		return
	}

	totalPrice := c.calculateTotalPrice(venue.VenueType, product.Price, installation.Price, room.Price, float64(params.RoomQuantity))

	insertOrder := order.Order{
		OrderNumber:    orderNumber,
		BuyerID:        userID.(string),
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
		Status:         4,
		CreatedBy:      userID.(string),
		LastUpdateBy:   userID.(string),
		ProjectID:      projectID,
		Email:          params.Email,
	}

	err = c.order.Insert(&insertOrder, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePostOrderByAgent] failed post order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post order", http.StatusInternalServerError)
		return
	}

	err = c.insertOrderDetail(insertOrder, device, product, installation, room, aging, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePostOrderByAgent] failed post order details, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post order details", http.StatusInternalServerError)
		return
	}

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
		_id       = router.GetParam(r, "id")
		id, err   = strconv.ParseInt(_id, 10, 64)
		params    reqUserID
		projectID = int64(10)
		isAdmin   = false
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
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handlePatchOrderForPayment] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	//check open payment status
	var getOrder order.Order
	if isAdmin {
		getOrder, err = c.order.Get(id, projectID, "")
	} else {
		getOrder, err = c.order.Get(id, projectID, userID.(string))
	}
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
		LastUpdateBy: userID.(string),
		PendingAt:    getOrder.PendingAt,
		PaidAt:       getOrder.PaidAt,
		FailedAt:     getOrder.FailedAt,
		VenueID:      getOrder.VenueID,
		BuyerID:      getOrder.BuyerID,
	}

	err = c.order.UpdateOrderStatus(&updateStatus, isAdmin)
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
		params    reqOrder
		_id       = router.GetParam(r, "id")
		id, err   = strconv.ParseInt(_id, 10, 64)
		projectID = int64(10)
		isAdmin   = false
	)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	//validasi request body
	err = form.Bind(&params, r)
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
		if params.UserID == "" {
			c.reporter.Errorf("[handlePatchOrder] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	//validasi order
	var getOrder order.Order
	if isAdmin {
		getOrder, err = c.order.Get(id, projectID, "")
	} else {
		getOrder, err = c.order.Get(id, projectID, userID.(string))
	}
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

	//validasi foreign key
	var venue venue.Venue
	if isAdmin {
		venue, err = c.venue.Get(projectID, params.VenueID, "")
	} else {
		venue, err = c.venue.Get(projectID, params.VenueID, userID.(string))
	}
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Venue Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue Not Found", http.StatusNotFound)
		return
	}

	device, err := c.device.Get(projectID, params.DeviceID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Device Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Device Not Found", http.StatusNotFound)
		return
	}

	product, err := c.product.Get(projectID, params.ProductID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Product Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Product Not Found", http.StatusNotFound)
		return
	}

	installation, err := c.installation.Get(params.InstallationID, projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Installation Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation Not Found", http.StatusNotFound)
		return
	}

	var room room.Room
	if params.RoomID != 0 && params.RoomQuantity != 0 {
		room, err = c.room.Get(projectID, params.RoomID)
		if err == sql.ErrNoRows {
			c.reporter.Errorf("[handlePostOrder] Room Not Found, err: %s", err.Error())
			view.RenderJSONError(w, "Room Not Found", http.StatusNotFound)
			return
		}
	}

	aging, err := c.aging.Get(params.AgingID, projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Aging Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging Not Found", http.StatusNotFound)
		return
	}

	valid := isOrderValid(venue.VenueType, venue.Capacity, params.AgingID, params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity)
	if !valid {
		c.reporter.Errorf("[handlePatchOrder] Order not valid, venueType: %d, capacity: %d, agingID: %d, deviceID: %d, productID: %d, installationID: %d, roomID: %d, roomQuantity: %d",
			venue.VenueType, venue.Capacity, params.AgingID, params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity)
		view.RenderJSONError(w, "Order not valid", http.StatusBadRequest)
		return
	}

	//validasi order detail
	if isAdmin {
		_, err = c.orderDetail.GetFromDBByOrderID(id, projectID, "")
	} else {
		_, err = c.orderDetail.GetFromDBByOrderID(id, projectID, userID.(string))
	}
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] order details not found, err: %s", err.Error())
		view.RenderJSONError(w, "Order details not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Failed get order details, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get order details", http.StatusInternalServerError)
		return
	}

	//calculate total price
	totalPrice := c.calculateTotalPrice(venue.VenueType, product.Price, installation.Price, room.Price, float64(params.RoomQuantity))

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
		ProjectID:      projectID,
		Status:         getOrder.Status,
		CreatedBy:      getOrder.CreatedBy,
		LastUpdateBy:   userID.(string),
		Email:          params.Email,
	}

	err = c.order.Update(&updateOrder, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] failed update order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update order", http.StatusInternalServerError)
		return
	}

	//update order details
	err = c.updateOrderDetail(updateOrder, device, product, installation, room, aging, isAdmin)
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
		params    reqUpdateOrderStatus
		_id       = router.GetParam(r, "id")
		id, err   = strconv.ParseInt(_id, 10, 64)
		projectID = int64(10)
		isAdmin   = false
	)
	if err != nil {
		c.reporter.Errorf("[handleUpdateOrderStatus] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handleUpdateOrderStatus] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleUpdateOrderStatus] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		if params.UserID == "" {
			c.reporter.Errorf("[handleUpdateOrderStatus] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	var getOrder order.Order
	if isAdmin {
		getOrder, err = c.order.Get(id, projectID, "")
	} else {
		getOrder, err = c.order.Get(id, projectID, userID.(string))
	}
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

	//update status order
	updateStatus := order.Order{
		OrderID:      id,
		ProjectID:    projectID,
		Status:       params.Status,
		CreatedBy:    getOrder.CreatedBy,
		LastUpdateBy: userID.(string),
		PendingAt:    getOrder.PendingAt,
		PaidAt:       getOrder.PaidAt,
		FailedAt:     getOrder.FailedAt,
		VenueID:      getOrder.VenueID,
		BuyerID:      getOrder.BuyerID,
	}

	err = c.order.UpdateOrderStatus(&updateStatus, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handleUpdateOrderStatus] failed update order status, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update order status", http.StatusInternalServerError)
		return
	}

	if updateStatus.Status == 2 {
		if isAdmin {
			userID = ""
		}

		result := c.handleEmailECert(getOrder.VenueID, userID.(string))
		if result == false {
			c.reporter.Warningf("[handleUpdateOrderStatus] Failed sent email ECert")
			view.RenderJSONError(w, "Failed sent email ECert", http.StatusInternalServerError)
			return
		}

		result = c.handleEmailInvoice(updateStatus.OrderID, userID.(string))
		if result == false {
			c.reporter.Warningf("[handleUpdateOrderStatus] Failed sent email invoice")
			view.RenderJSONError(w, "Failed sent email invoice", http.StatusInternalServerError)
			return
		}
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
		params    reqUpdateOpenPaymentStatus
		_id       = router.GetParam(r, "id")
		id, err   = strconv.ParseInt(_id, 10, 64)
		projectID = int64(10)
		isAdmin   = false
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
		isAdmin = true
	}

	var getOrder order.Order
	if isAdmin {
		getOrder, err = c.order.Get(id, projectID, "")
	} else {
		getOrder, err = c.order.Get(id, projectID, userID.(string))
	}
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
		ProjectID:         projectID,
		OpenPaymentStatus: params.OpenPaymentStatus,
		CreatedBy:         getOrder.CreatedBy,
		LastUpdateBy:      userID.(string),
		VenueID:           getOrder.VenueID,
		BuyerID:           getOrder.BuyerID,
	}

	err = c.order.UpdateOpenPaymentStatus(&updateStatus, isAdmin)
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
		_id       = router.GetParam(r, "id")
		id, err   = strconv.ParseInt(_id, 10, 64)
		params    reqUserID
		projectID = int64(10)
		isAdmin   = false
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
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handleDeleteOrder] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	//validasi order
	var getOrder order.Order
	if isAdmin {
		getOrder, err = c.order.Get(id, projectID, "")
	} else {
		getOrder, err = c.order.Get(id, projectID, userID.(string))
	}
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
	if isAdmin {
		_, err = c.orderDetail.GetFromDBByOrderID(id, projectID, "")
	} else {
		_, err = c.orderDetail.GetFromDBByOrderID(id, projectID, userID.(string))
	}
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

	//delete order
	deleteOrder := order.Order{
		OrderID:      id,
		ProjectID:    projectID,
		CreatedBy:    getOrder.CreatedBy,
		LastUpdateBy: userID.(string),
		BuyerID:      getOrder.BuyerID,
		VenueID:      getOrder.VenueID,
		Status:       getOrder.Status,
		PaidAt:       getOrder.PaidAt,
	}

	err = c.order.Delete(&deleteOrder, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handleDeleteOrder] failed delete order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete order", http.StatusInternalServerError)
		return
	}

	//delete order details
	err = c.deleteOrderDetail(deleteOrder, isAdmin)
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
	var projectID = int64(10)

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetAllOrder] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		userID = ""
	}

	orders, err := c.order.Select(projectID, userID.(string))
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
		_id       = router.GetParam(r, "id")
		id, err   = strconv.ParseInt(_id, 10, 64)
		projectID = int64(10)
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

	order, err := c.order.Get(id, projectID, userID.(string))
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

func (c *Controller) handleGetAllByVenueID(w http.ResponseWriter, r *http.Request) {
	var (
		_venueID     = router.GetParam(r, "venue_id")
		venueID, err = strconv.ParseInt(_venueID, 10, 64)
		projectID    = int64(10)
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
		userID = ""
	}

	orders, err := c.order.SelectByVenueID(venueID, projectID, userID.(string))
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
		buyerID   = router.GetParam(r, "buyer_id")
		projectID = int64(10)
	)

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetAllOrdersByBuyerID] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		userID = ""
	}

	orders, err := c.order.SelectByBuyerID(buyerID, projectID, userID.(string))
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
		projectID = int64(10)
	)

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetAllOrdersByPaidDate] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		userID = ""
	}

	orders, err := c.order.SelectByPaidDate(paidDate, projectID, userID.(string))
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

func (c *Controller) handleCalculateOrderPrice(w http.ResponseWriter, r *http.Request) {
	var (
		projectID = int64(10)
		params    reqCalculateOrderPrice
	)

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Errorf("[handleCalculateOrderPrice] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleCalculateOrderPrice] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		userID = ""
	}

	venue, err := c.venue.Get(projectID, params.VenueID, userID.(string))
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleCalculateOrderPrice] Venue Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue Not Found", http.StatusNotFound)
		return
	}

	device, err := c.device.Get(projectID, params.DeviceID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleCalculateOrderPrice] Device Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Device Not Found", http.StatusNotFound)
		return
	}

	product, err := c.product.Get(projectID, params.ProductID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleCalculateOrderPrice] Product Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Product Not Found", http.StatusNotFound)
		return
	}

	installation, err := c.installation.Get(params.InstallationID, projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleCalculateOrderPrice] Installation Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Installation Not Found", http.StatusNotFound)
		return
	}

	var room room.Room
	if params.RoomID != 0 && params.RoomQuantity != 0 {
		room, err = c.room.Get(projectID, params.RoomID)
		if err == sql.ErrNoRows {
			c.reporter.Errorf("[handleCalculateOrderPrice] Room Not Found, err: %s", err.Error())
			view.RenderJSONError(w, "Room Not Found", http.StatusNotFound)
			return
		}
	}

	aging, err := c.aging.Get(params.AgingID, projectID)
	if err == sql.ErrNoRows {
		c.reporter.Errorf("[handleCalculateOrderPrice] Aging Not Found, err: %s", err.Error())
		view.RenderJSONError(w, "Aging Not Found", http.StatusNotFound)
		return
	}

	valid := isOrderValid(venue.VenueType, venue.Capacity, params.AgingID, params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity)
	if !valid {
		c.reporter.Errorf("[handleCalculateOrderPrice] Order not valid, venueType: %d, capacity: %d, agingID: %d, deviceID: %d, productID: %d, installationID: %d, roomID: %d, roomQuantity: %d",
			venue.VenueType, venue.Capacity, params.AgingID, params.DeviceID, params.ProductID, params.InstallationID, params.RoomID, params.RoomQuantity)
		view.RenderJSONError(w, "Order not valid", http.StatusBadRequest)
		return
	}

	//calculate total price
	totalPrice := c.calculateTotalPrice(venue.VenueType, product.Price, installation.Price, room.Price, float64(params.RoomQuantity))

	//set response
	details := c.mappingDetailOrder(params.RoomQuantity, device, product, installation, room, aging)

	res := view.CalculatePriceAttributes{
		TotalPrice: totalPrice,
		Details:    details,
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func leftPadLen(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

func isOrderValid(venueType, venueCapacity, agingID, deviceID, productID, installationID, roomID, roomQuantity int64) bool {
	if roomID == 0 && roomQuantity == 0 {
		if venueType > 0 && venueType <= 4 {
			if venueCapacity == 1 {
				if agingID == 1 {
					if deviceID == 1 {
						if productID == 1 && installationID == 1 {
							return true
						}
					} else if deviceID == 2 {
						if productID == 4 && installationID == 4 {
							return true
						}
					} else if deviceID == 3 {
						if productID == 9 && installationID == 9 {
							return true
						}
					} else if deviceID == 4 {
						if productID == 10 && installationID == 10 {
							return true
						}
					}
				} else if agingID == 2 {
					if deviceID == 1 && productID == 2 && installationID == 2 {
						return true
					}
				} else if agingID == 3 {
					if deviceID == 1 && productID == 3 && installationID == 3 {
						return true
					}
				}
			} else if venueCapacity == 2 && agingID == 1 {
				if deviceID == 1 {
					if productID == 5 && installationID == 5 {
						return true
					}
				} else if deviceID == 2 {
					if productID == 6 && installationID == 6 {
						return true
					}
				} else if deviceID == 3 {
					if productID == 11 && installationID == 11 {
						return true
					}
				} else if deviceID == 4 {
					if productID == 12 && installationID == 12 {
						return true
					}
				}
			}
		} else if venueType == 5 || venueType == 6 {
			if venueCapacity == 1 {
				if agingID == 1 {
					if deviceID == 1 {
						if productID == 1 && installationID == 1 {
							return true
						}
					} else if deviceID == 2 {
						if productID == 4 && installationID == 4 {
							return true
						}
					} else if deviceID == 3 {
						if productID == 9 && installationID == 9 {
							return true
						}
					} else if deviceID == 4 {
						if productID == 10 && installationID == 10 {
							return true
						}
					}
				} else if agingID == 2 {
					if deviceID == 1 && productID == 2 && installationID == 2 {
						return true
					}
				} else if agingID == 3 {
					if deviceID == 1 && productID == 3 && installationID == 3 {
						return true
					}
				}
			}
		} else if venueType >= 7 && venueType <= 11 {
			if venueCapacity == 2 && agingID == 1 {
				if deviceID == 1 {
					if productID == 5 && installationID == 5 {
						return true
					}
				} else if deviceID == 2 {
					if productID == 6 && installationID == 6 {
						return true
					}
				} else if deviceID == 3 {
					if productID == 11 && installationID == 11 {
						return true
					}
				} else if deviceID == 4 {
					if productID == 12 && installationID == 12 {
						return true
					}
				}
			}
		}
	} else {
		if venueCapacity == 0 && agingID == 1 {
			if venueType >= 12 && venueType <= 14 {
				if roomID == 1 {
					if deviceID == 1 {
						if productID == 7 && installationID == 7 {
							return true
						}
					} else if deviceID == 3 {
						if productID == 13 && installationID == 13 {
							return true
						}
					}
				}
			} else if venueType >= 15 && venueType <= 18 {
				if roomID == 2 {
					if deviceID == 1 {
						if productID == 8 && installationID == 8 {
							return true
						}
					} else if deviceID == 3 {
						if productID == 14 && installationID == 14 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (c *Controller) calculateTotalPrice(venueType int64, productPrice, installationPrice, roomPrice, roomQuantity float64) float64 {
	var totalPrice float64
	if venueType > 0 && venueType <= 11 {
		totalPrice = (productPrice + installationPrice)
	} else if venueType > 11 && venueType <= 18 {
		totalPrice = (installationPrice + (math.Ceil(roomQuantity*0.3) * roomPrice))
	}

	return totalPrice

}

func (c *Controller) generateOrderNumber() (string, error) {
	lastOrderNumber, err := c.order.GetLastOrderNumber()
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	dateNow := time.Now().Format("060102")
	if strings.Compare(dateNow, lastOrderNumber.Date) == 1 {
		lastOrderNumber.Number = 0
	}
	return "MN" + dateNow + leftPadLen(strconv.FormatInt((lastOrderNumber.Number+1), 10), "0", 7), nil
}
