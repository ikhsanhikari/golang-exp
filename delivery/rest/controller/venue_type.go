package controller

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
	"fmt"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue_type"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllVenueTypes(w http.ResponseWriter, r *http.Request) {
	venueTypes, err := c.venueType.Select(c.projectID)
	if err != nil {
		c.reporter.Errorf("[handleGetAllVenueTypes] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get VenueTypes", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(venueTypes))
	for _, venueType := range venueTypes {
		res = append(res, view.DataResponse{
			Type: "venueTypes",
			ID:   venueType.Id,
			Attributes: view.VenueTypeAttributes{
				Id:               venueType.Id,
				Name:             venueType.Name,
				Description:      venueType.Description,
				Capacity:         venueType.Capacity,
				CommercialTypeID: venueType.CommercialTypeID,
				PricingGroupID:   venueType.PricingGroupID,
				CreatedAt:        venueType.CreatedAt,
				UpdatedAt:        venueType.UpdatedAt,
				DeletedAt:        venueType.DeletedAt,
				Status:           venueType.Status,
				ProjectID:        venueType.ProjectID,
				CreatedBy:        venueType.CreatedBy,
				LastUpdateBy:     venueType.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (c *Controller) handleDeleteVenueType(w http.ResponseWriter, r *http.Request) {
	var (
		params  reqDeleteVenueType
		isAdmin = false
	)

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteAging] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handlePostAging] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteVenueType] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	venueTy, err := c.venueType.Get(c.projectID, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteVenueType] VenueType not found, err: %s", err.Error())
		view.RenderJSONError(w, "VenueType not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteVenueType] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get VenueType", http.StatusInternalServerError)
		return
	}

	err = c.venueType.Delete(c.projectID, id, venueTy.CommercialTypeID, userID.(string), isAdmin)
	if err != nil {
		c.reporter.Errorf("[handleDeleteVenueType] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete VenueType", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostVenueType(w http.ResponseWriter, r *http.Request) {
	var params reqVenueType
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostVenueType] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	venueType := venue_type.VenueType{
		Id:               params.Id,
		Name:             params.Name,
		Description:      params.Description,
		Capacity:         params.Capacity,
		CommercialTypeID: params.CommercialTypeID,
		PricingGroupID:   params.PricingGroupID,
		CreatedBy:        params.CreatedBy,
		CreatedAt:		  time.Now(),
	    UpdatedAt:		  time.Now(),
	    Status:			  1,
		LastUpdateBy:    params.CreatedBy,
		ProjectID:		 c.projectID,
	}

	err = c.venueType.Insert(&venueType)
	if err != nil {
		c.reporter.Infof("[handlePostVenueType] error insert VenueType repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post VenueType", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, venueType, http.StatusOK)
}

func (c *Controller) handlePatchVenueType(w http.ResponseWriter, r *http.Request) {
	var (
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)

	if err != nil {
		c.reporter.Warningf("[handlePatchVenueType] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqVenueType
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchVenueType] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePatchAging] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		if params.LastUpdateBy == "" {
			c.reporter.Errorf("[handlePatchAging] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.LastUpdateBy
		isAdmin = true
	}

	venueTy, err := c.venueType.Get(c.projectID, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchVenueType] VenueType not found, err: %s", err.Error())
		view.RenderJSONError(w, "VenueType not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchVenueType] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get VenueType", http.StatusInternalServerError)
		return
	}
	venueType := venue_type.VenueType{
		Id:               id,
		Name:             params.Name,
		Description:      params.Description,
		Capacity:         params.Capacity,
		CommercialTypeID: params.CommercialTypeID,
		PricingGroupID:   params.PricingGroupID,
		LastUpdateBy:     userID.(string),
		ProjectID:		  c.projectID,
		UpdatedAt:		  time.Now(),
	}
	err = c.venueType.Update(&venueType, venueTy.CommercialTypeID, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePatchVenueType] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update VenueType", http.StatusInternalServerError)
		return
	}
	res := view.DataResponse{
		ID:   id,
		Type: "venueTypes",
		Attributes: view.VenueTypeAttributes{
			Id:               venueType.Id,
			Name:             venueType.Name,
			Description:      venueType.Description,
			Capacity:         venueType.Capacity,
			PricingGroupID:   venueType.PricingGroupID,
			CommercialTypeID: venueType.CommercialTypeID,
			CreatedAt:        venueType.CreatedAt,
			UpdatedAt:        venueType.UpdatedAt,
			DeletedAt:        venueType.DeletedAt,
			Status:           venueType.Status,
			ProjectID:        venueType.ProjectID,
			CreatedBy:        venueType.CreatedBy,
			LastUpdateBy:     venueType.LastUpdateBy,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetVenueTypeByCommercialTypeID(w http.ResponseWriter, r *http.Request) {
	ctid, err := strconv.ParseInt(router.GetParam(r, "commercialTypeId"), 10, 64)

	if err != nil {
		c.reporter.Errorf("[handleGetVenueTypeByCommercialTypeID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	venueTypes, err := c.venueType.GetByCommercialType(c.projectID, ctid)
	if err != nil {
		c.reporter.Errorf("[handleGetVenueTypeByCommercialTypeID] order not found, err: %s", err.Error())
		view.RenderJSONError(w, "Orders not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetVenueTypeByCommercialTypeID] failed get order, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get orders", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(venueTypes))
	for _, venueType := range venueTypes {
		res = append(res, view.DataResponse{
			Type: "venueTypes",
			ID:   venueType.Id,
			Attributes: view.VenueTypeAttributes{
				Id:               venueType.Id,
				Name:             venueType.Name,
				Description:      venueType.Description,
				Capacity:         venueType.Capacity,
				CommercialTypeID: venueType.CommercialTypeID,
				PricingGroupID:   venueType.PricingGroupID,
				CreatedAt:        venueType.CreatedAt,
				UpdatedAt:        venueType.UpdatedAt,
				DeletedAt:        venueType.DeletedAt,
				Status:           venueType.Status,
				ProjectID:        venueType.ProjectID,
				CreatedBy:        venueType.CreatedBy,
				LastUpdateBy:     venueType.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}
