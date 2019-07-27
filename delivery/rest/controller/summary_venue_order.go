package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetSumOrderByID(w http.ResponseWriter, r *http.Request) {
	var (
		_id       = router.GetParam(r, "id")
		id, err   = strconv.ParseInt(_id, 10, 64)
		projectID = int64(10)
	)
	if err != nil {
		c.reporter.Errorf("[handleGetSumOrderByVenueID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetSumOrderByVenueID] failed get user")
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

	sumvenue, err := c.order.GetSummaryVenueByVenueID(id, projectID, fmt.Sprintf("%v", userID))
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetSumOrderByVenueID] failed get sum venue, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get sum venue", http.StatusInternalServerError)
		return
	}

	sumorders, err := c.order.SelectSummaryOrdersByVenueID(id, projectID, fmt.Sprintf("%v", userID))
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetSumOrderByVenueID] failed get sum orders, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get sum orders", http.StatusInternalServerError)
		return
	}

	orders := make([]view.SumOrderAttributes, 0, len(sumorders))
	for _, sumorder := range sumorders {
		orders = append(orders, view.SumOrderAttributes{
			OrderID:           sumorder.OrderID,
			OrderNumber:       sumorder.OrderNumber,
			OrderTotalPrice:   sumorder.OrderTotalPrice,
			OrderCreatedAt:    sumorder.OrderCreatedAt,
			OrderPaidAt:       sumorder.OrderPaidAt,
			OrderFailedAt:     sumorder.OrderFailedAt,
			OrderEmail:        sumorder.OrderEmail,
			DeviceName:        sumorder.DeviceName,
			ProductName:       sumorder.ProductName,
			InstallationName:  sumorder.InstallationName,
			RoomName:          sumorder.RoomName,
			RoomQty:           sumorder.RoomQty,
			AgingName:         sumorder.AgingName,
			OrderStatus:       sumorder.OrderStatus,
			OpenPaymentStatus: sumorder.OpenPaymentStatus,
			EcertLastSentDate: sumorder.EcertLastSentDate},
		)
	}

	var res view.DataResponseOrder
	if sumvenue.VenueID != 0 {
		res = view.DataResponseOrder{
			ID:   sumvenue.VenueID,
			Type: "summary_venue_order",
			Attributes: view.SumVenueAttributes{
				VenueName:             sumvenue.VenueName,
				VenueType:             sumvenue.VenueType,
				VenuePhone:            sumvenue.VenuePhone,
				VenuePicName:          sumvenue.VenuePicName,
				VenuePicContactNumber: sumvenue.VenuePicContactNumber,
				VenueAddress:          sumvenue.VenueAddress,
				VenueCity:             sumvenue.VenueCity,
				VenueProvince:         sumvenue.VenueProvince,
				VenueZip:              sumvenue.VenueZip,
				VenueCapacity:         sumvenue.VenueCapacity,
				VenueFacilities:       sumvenue.VenueFacilities,
				VenueLongitude:        sumvenue.VenueLongitude,
				VenueLatitude:         sumvenue.VenueLatitude,
				VenueCategory:         sumvenue.VenueCategory,
				VenueShowStatus:       sumvenue.VenueShowStatus,
				CompanyID:             sumvenue.CompanyID,
				CompanyName:           sumvenue.CompanyName,
				CompanyAddress:        sumvenue.CompanyAddress,
				CompanyCity:           sumvenue.CompanyCity,
				CompanyProvince:       sumvenue.CompanyProvince,
				CompanyZip:            sumvenue.CompanyZip,
				CompanyEmail:          sumvenue.CompanyEmail,
				EcertLastSent:         sumvenue.EcertLastSent,
				LicenseNumber:         sumvenue.LicenseNumber,
				LicenseActiveDate:     sumvenue.LicenseActiveDate,
				LicenseExpiredDate:    sumvenue.LicenseExpiredDate,
				LastOrderID:           sumvenue.LastOrderID,
				LastOrderNumber:       sumvenue.LastOrderNumber,
				LastOrderTotalPrice:   sumvenue.LastOrderTotalPrice,
				LastRoomID:            sumvenue.LastRoomID,
				LastRoomQuantity:      sumvenue.LastRoomQuantity,
				LastAgingID:           sumvenue.LastAgingID,
				LastDeviceID:          sumvenue.LastDeviceID,
				LastProductID:         sumvenue.LastProductID,
				LastInstallationID:    sumvenue.LastInstallationID,
				LastOrderCreatedAt:    sumvenue.LastOrderCreatedAt,
				LastOrderPaidAt:       sumvenue.LastOrderPaidAt,
				LastOrderFailedAt:     sumvenue.LastOrderFailedAt,
				LastOrderEmail:        sumvenue.LastOrderEmail,
				LastOrderStatus:       sumvenue.LastOrderStatus,
				Orders:                orders,
			},
		}
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetLicenseByIDForChecker(w http.ResponseWriter, r *http.Request) {
	var (
		_id       = router.GetParam(r, "id")
		projectID = int64(10)
	)

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetLicenseByIDForChecker] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleGetLicenseByIDForChecker] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}
	// harus diganti dengan pengecekan user checker
	fmt.Println(userID)

	sumvenue, err := c.order.GetSummaryVenueByLicenseNumber(_id, projectID)
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetLicenseByIDForChecker] failed get sum venue, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get sum venue", http.StatusInternalServerError)
		return
	}

	sumorders, err := c.order.SelectSummaryOrdersByLicenseNumber(_id, projectID)
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetLicenseByIDForChecker] failed get sum orders, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get sum orders", http.StatusInternalServerError)
		return
	}

	orders := make([]view.SumOrderAttributes, 0, len(sumorders))
	for _, sumorder := range sumorders {
		orders = append(orders, view.SumOrderAttributes{
			OrderID:           sumorder.OrderID,
			OrderNumber:       sumorder.OrderNumber,
			OrderTotalPrice:   sumorder.OrderTotalPrice,
			OrderCreatedAt:    sumorder.OrderCreatedAt,
			OrderPaidAt:       sumorder.OrderPaidAt,
			OrderFailedAt:     sumorder.OrderFailedAt,
			OrderEmail:        sumorder.OrderEmail,
			DeviceName:        sumorder.DeviceName,
			ProductName:       sumorder.ProductName,
			InstallationName:  sumorder.InstallationName,
			RoomName:          sumorder.RoomName,
			RoomQty:           sumorder.RoomQty,
			AgingName:         sumorder.AgingName,
			OrderStatus:       sumorder.OrderStatus,
			OpenPaymentStatus: sumorder.OpenPaymentStatus,
			EcertLastSentDate: sumorder.EcertLastSentDate},
		)
	}

	var res view.DataResponseOrder
	if sumvenue.VenueID != 0 {
		res = view.DataResponseOrder{
			ID:   sumvenue.VenueID,
			Type: "summary_venue_order",
			Attributes: view.SumVenueAttributes{
				VenueName:          sumvenue.VenueName,
				VenueType:          sumvenue.VenueType,
				VenueAddress:       sumvenue.VenueAddress,
				VenueCity:          sumvenue.VenueCity,
				VenueProvince:      sumvenue.VenueProvince,
				VenueZip:           sumvenue.VenueZip,
				VenueCapacity:      sumvenue.VenueCapacity,
				VenueLongitude:     sumvenue.VenueLongitude,
				VenueLatitude:      sumvenue.VenueLatitude,
				VenueCategory:      sumvenue.VenueCategory,
				VenueShowStatus:    sumvenue.VenueShowStatus,
				CompanyName:        sumvenue.CompanyName,
				CompanyAddress:     sumvenue.CompanyAddress,
				CompanyCity:        sumvenue.CompanyCity,
				CompanyProvince:    sumvenue.CompanyProvince,
				CompanyZip:         sumvenue.CompanyZip,
				CompanyEmail:       sumvenue.CompanyEmail,
				LicenseNumber:      sumvenue.LicenseNumber,
				LicenseActiveDate:  sumvenue.LicenseActiveDate,
				LicenseExpiredDate: sumvenue.LicenseExpiredDate,
				Orders:             orders,
			},
		}
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetSumOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	projectID := int64(10)

	getParam := r.URL.Query()
	limitVal := getParam.Get("limit")
	offsetVal := getParam.Get("page")
	pagination := getParam.Get("pagination")
	var err error
	var sumvenues order.SummaryVenues
	offset := 1
	limit := 15
	if limitVal != "" {
		limit, err = strconv.Atoi(limitVal)
	}
	if offsetVal != "" {
		offset, err = strconv.Atoi(offsetVal)
	}
	offset = offset - 1
	offset = limit * offset
	limit = limit + 1

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
	if pagination == "true" {
		sumvenues, err = c.order.SelectSummaryVenuesByUserID(projectID, fmt.Sprintf("%v", userID))
		if err != nil && err != sql.ErrNoRows {
			c.reporter.Errorf("[handleGetSumOrdersByUserIDPagination] failed get sum venues, err: %s", err.Error())
			view.RenderJSONError(w, "Failed get sum venues", http.StatusInternalServerError)
			return
		}
	} else {
		sumvenues, err = c.order.SelectSummaryVenuesByUserID(projectID, fmt.Sprintf("%v", userID))
		if err != nil && err != sql.ErrNoRows {
			c.reporter.Errorf("[handleGetSumOrdersByUserID] failed get sum venues, err: %s", err.Error())
			view.RenderJSONError(w, "Failed get sum venues", http.StatusInternalServerError)
			return
		}
	}

	res := make([]view.DataResponseOrder, 0, len(sumvenues))
	for _, sumvenue := range sumvenues {

		sumorders, err := c.order.SelectSummaryOrdersByVenueID(sumvenue.VenueID, projectID, fmt.Sprintf("%v", userID))
		if err != nil && err != sql.ErrNoRows {
			c.reporter.Errorf("[handleGetSumOrdersByUserID] failed get sum order, err: %s", err.Error())
			view.RenderJSONError(w, "Failed get sum orders", http.StatusInternalServerError)
			return
		}

		orders := make([]view.SumOrderAttributes, 0, len(sumorders))
		for _, sumorder := range sumorders {
			orders = append(orders, view.SumOrderAttributes{
				OrderID:           sumorder.OrderID,
				OrderNumber:       sumorder.OrderNumber,
				OrderTotalPrice:   sumorder.OrderTotalPrice,
				OrderCreatedAt:    sumorder.OrderCreatedAt,
				OrderPaidAt:       sumorder.OrderPaidAt,
				OrderFailedAt:     sumorder.OrderFailedAt,
				OrderEmail:        sumorder.OrderEmail,
				DeviceName:        sumorder.DeviceName,
				ProductName:       sumorder.ProductName,
				InstallationName:  sumorder.InstallationName,
				RoomName:          sumorder.RoomName,
				RoomQty:           sumorder.RoomQty,
				AgingName:         sumorder.AgingName,
				OrderStatus:       sumorder.OrderStatus,
				OpenPaymentStatus: sumorder.OpenPaymentStatus,
				EcertLastSentDate: sumorder.EcertLastSentDate},
			)
		}

		res = append(res, view.DataResponseOrder{
			ID:   sumvenue.VenueID,
			Type: "summary_venue_order",
			Attributes: view.SumVenueAttributes{
				VenueName:             sumvenue.VenueName,
				VenueType:             sumvenue.VenueType,
				VenuePhone:            sumvenue.VenuePhone,
				VenuePicName:          sumvenue.VenuePicName,
				VenuePicContactNumber: sumvenue.VenuePicContactNumber,
				VenueAddress:          sumvenue.VenueAddress,
				VenueCity:             sumvenue.VenueCity,
				VenueProvince:         sumvenue.VenueProvince,
				VenueZip:              sumvenue.VenueZip,
				VenueCapacity:         sumvenue.VenueCapacity,
				VenueFacilities:       sumvenue.VenueFacilities,
				VenueLongitude:        sumvenue.VenueLongitude,
				VenueLatitude:         sumvenue.VenueLatitude,
				VenueCategory:         sumvenue.VenueCategory,
				VenueShowStatus:       sumvenue.VenueShowStatus,
				CompanyID:             sumvenue.CompanyID,
				CompanyName:           sumvenue.CompanyName,
				CompanyAddress:        sumvenue.CompanyAddress,
				CompanyCity:           sumvenue.CompanyCity,
				CompanyProvince:       sumvenue.CompanyProvince,
				CompanyZip:            sumvenue.CompanyZip,
				CompanyEmail:          sumvenue.CompanyEmail,
				EcertLastSent:         sumvenue.EcertLastSent,
				LicenseNumber:         sumvenue.LicenseNumber,
				LicenseActiveDate:     sumvenue.LicenseActiveDate,
				LicenseExpiredDate:    sumvenue.LicenseExpiredDate,
				LastOrderID:           sumvenue.LastOrderID,
				LastOrderNumber:       sumvenue.LastOrderNumber,
				LastOrderTotalPrice:   sumvenue.LastOrderTotalPrice,
				LastRoomID:            sumvenue.LastRoomID,
				LastRoomQuantity:      sumvenue.LastRoomQuantity,
				LastAgingID:           sumvenue.LastAgingID,
				LastDeviceID:          sumvenue.LastDeviceID,
				LastProductID:         sumvenue.LastProductID,
				LastInstallationID:    sumvenue.LastInstallationID,
				LastOrderCreatedAt:    sumvenue.LastOrderCreatedAt,
				LastOrderPaidAt:       sumvenue.LastOrderPaidAt,
				LastOrderFailedAt:     sumvenue.LastOrderFailedAt,
				LastOrderEmail:        sumvenue.LastOrderEmail,
				LastOrderStatus:       sumvenue.LastOrderStatus,
				Orders:                orders,
			},
		})
	}

	limit = limit - 1
	var hasNext bool
	hasNext = false
	if len(res) > limit {
		hasNext = true
		//view.RenderJSONDataPage(w, res, hasNext, http.StatusOK)
	}
	//else {
	//view.RenderJSONData(w, res, http.StatusOK)
	//}
	view.RenderJSONDataPage(w, res, hasNext, http.StatusOK)
}
