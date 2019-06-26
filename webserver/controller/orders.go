package controller

import (
	"log"
	"net/http"
	"strconv"

	//"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/orders"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/webserver/view"
	"github.com/julienschmidt/httprouter"
)

func (h *handler) handleGetAllByVenueID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		venue = ps.ByName("venue_id")
		//project, _ = authpassport.GetProject(r)
		//pid        = project.ID
		// resVid     responseVideo
	)

	venue_id, _ := strconv.Atoi(venue)
	orders, err := h.orders.SelectByVenueId(venue_id)
	if err != nil {
		log.Println(err)
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

func (h *handler) handleGetAllByBuyerID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		buyer = ps.ByName("buyer_id")
		//project, _ = authpassport.GetProject(r)
		//pid        = project.ID
		// resVid     responseVideo
	)

	buyer_id, _ := strconv.Atoi(buyer)
	orders, err := h.orders.SelectByVenueId(buyer_id)
	if err != nil {
		log.Println(err)
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
func (h *handler) handleGetAllByPaidDate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		paiddate = ps.ByName("paid_date")
		//project, _ = authpassport.GetProject(r)
		//pid        = project.ID
		// resVid     responseVideo
	)

	orders, err := h.orders.SelectByPaidDate(paiddate)
	if err != nil {
		log.Println(err)
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
