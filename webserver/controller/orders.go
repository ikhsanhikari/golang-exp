package controller

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"

	// "git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/orders"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/webserver/view"
)

func (h *handler) handlePostOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		// project, _ = authpassport.GetProject(r)
		//  pid        = project.ID
		params reqOrderInsert
	)

	err := form.Bind(&params, r)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	num, err := h.orders.GetID()
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Failed get id", http.StatusInternalServerError)
		return
	}
	orderNumber := "MN" + time.Now().Format("060102") + leftPadLen(strconv.FormatInt(num+1, 10), "0", 7)

	order := orders.Order{
		OrderNumber:     orderNumber,
		BuyerID:         params.BuyerID,
		VenueID:         params.VenueID,
		ProductID:       params.ProductID,
		Quantity:        params.Quantity,
		TotalPrice:      params.TotalPrice,
		PaymentMethodID: params.PaymentMethodID,
		PaymentFee:      params.PaymentFee,
		ProjectID:       2,
	}

	err = h.orders.Insert(&order)
	if err != nil {
		log.Println(err)
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

func (h *handler) handlePatchOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		// project, _ = authpassport.GetProject(r)
		// 2        = project.ID
		params  reqOrderUpdate
		_id     = ps.ByName("id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = h.orders.Get(id, 2)
	if err == sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	order := orders.Order{
		OrderID:         id,
		BuyerID:         params.BuyerID,
		VenueID:         params.VenueID,
		ProductID:       params.ProductID,
		Quantity:        params.Quantity,
		TotalPrice:      params.TotalPrice,
		PaymentMethodID: params.PaymentMethodID,
		PaymentFee:      params.PaymentFee,
		ProjectID:       2,
		Status:          params.Status,
	}

	err = h.orders.Update(&order)
	if err != nil {
		log.Println(err)
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

func (h *handler) handleDeleteOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		_id     = ps.ByName("id")
		id, err = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = h.orders.Get(id, 2)
	if err == sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
		return
	}

	err = h.orders.Delete(id, 2)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Failed delete order", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (h *handler) handleGetOrders(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
	// project, _ = authpassport.GetProject(r)
	// pid        = project.ID
	)

	orders, err := h.orders.Select(2)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Orders not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Failed get orders", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponseOrders, 0, len(orders))
	for _, order := range orders {
		res = append(res, view.DataResponseOrders{
			ID:   order.OrderID,
			Type: "orders",
			Attributes: view.GetOrdersAttributes{
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
