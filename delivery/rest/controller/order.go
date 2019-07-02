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

	order := order.Order{
		OrderNumber:     orderNumber,
		BuyerID:         params.BuyerID,
		VenueID:         params.VenueID,
		ProductID:       params.ProductID,
		Quantity:        params.Quantity,
		TotalPrice:      params.TotalPrice,
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
		Type: "orders",
		Attributes: view.PostOrderAttributes{
			BuyerID:         order.BuyerID,
			VenueID:         order.VenueID,
			ProductID:       order.ProductID,
			Quantity:        order.Quantity,
			TotalPrice:      order.TotalPrice,
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

	_, err = c.order.Get(id, 10)
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

	order := order.Order{
		OrderID:         id,
		BuyerID:         params.BuyerID,
		VenueID:         params.VenueID,
		ProductID:       params.ProductID,
		Quantity:        params.Quantity,
		TotalPrice:      params.TotalPrice,
		PaymentMethodID: params.PaymentMethodID,
		PaymentFee:      params.PaymentFee,
		ProjectID:       10,
		Status:          params.Status,
	}

	err = c.order.Update(&order)
	if err != nil {
		c.reporter.Errorf("[handlePatchOrder] failed update order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update order", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		ID:   order.OrderID,
		Type: "orders",
		Attributes: view.PatchOrderAttributes{
			BuyerID:         order.BuyerID,
			VenueID:         order.VenueID,
			ProductID:       order.ProductID,
			Quantity:        order.Quantity,
			TotalPrice:      order.TotalPrice,
			PaymentMethodID: order.PaymentMethodID,
			PaymentFee:      order.PaymentFee,
			Status:          order.Status,
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

	view.RenderJSONData(w, "OK", http.StatusOK)
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
			Attributes: view.GetOrderAttributes{
				OrderNumber:     order.OrderNumber,
				BuyerID:         order.BuyerID,
				VenueID:         order.VenueID,
				ProductID:       order.ProductID,
				Quantity:        order.Quantity,
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
