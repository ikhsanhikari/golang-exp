package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue_type"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	auth "git.sstv.io/lib/go/go-auth-api.git/authpassport"
)

func (c *Controller) handleGetAllVenueTypes(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUser(r)
    if !ok {
		view.RenderJSONError(w, "Failed get User for VenueTypes", http.StatusInternalServerError)
		return
    }
   _, ok = user["sub"]
   if !ok {
		c.reporter.Errorf("[handleGetAllVenueTypes] error get IDUser")
   }
	venueTypes, err := c.venueType.Select(10)
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
				Id							:  venueType.Id,
				Name						:  venueType.Name,
				Description					:  venueType.Description,
				Capacity					:  venueType.Capacity,
				CommercialTypeID			:  venueType.CommercialTypeID,
				PricingGroupID				:  venueType.PricingGroupID,
				CreatedAt					:  venueType.CreatedAt,
				UpdatedAt					:  venueType.UpdatedAt,
				DeletedAt					:  venueType.DeletedAt,
				Status						:  venueType.Status,
				ProjectID					:  venueType.ProjectID,
				CreatedBy					:  venueType.CreatedBy,
				LastUpdateBy				:  venueType.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (c *Controller) handleDeleteVenueType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteVenueType] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.venueType.Get(10,id)
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

	err = c.venueType.Delete(10,id)
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
		Id							:  params.Id,
		Name						:  params.Name,
		Description					:  params.Description,
		Capacity					:  params.Capacity,
		CommercialTypeID			:  params.CommercialTypeID,
		PricingGroupID				:  params.PricingGroupID,
		CreatedBy					:  params.CreatedBy,
				
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
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
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

	_, err = c.venueType.Get(10,id)
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
		Id							:  params.Id,
		Name						:  params.Name,
		Description					:  params.Description,
		Capacity					:  params.Capacity,
		CommercialTypeID			:  params.CommercialTypeID,
		PricingGroupID				:  params.PricingGroupID,
		CreatedBy					:  params.CreatedBy,
				
	}
	err = c.venueType.Update(&venueType)
	if err != nil {
		c.reporter.Errorf("[handlePatchVenueType] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update VenueType", http.StatusInternalServerError)
		return
	}
	res := view.DataResponse{
		ID:   id,
		Type: "venueTypes",
		Attributes: view.VenueTypeAttributes{
			Id							:  venueType.Id,
			Name						:  venueType.Name,
			Description					:  venueType.Description,
			Capacity					:  venueType.Capacity,
			PricingGroupID				:  venueType.PricingGroupID,
			CreatedAt					:  venueType.CreatedAt,
			UpdatedAt					:  venueType.UpdatedAt,
			DeletedAt					:  venueType.DeletedAt,
			Status						:  venueType.Status,
			ProjectID					:  venueType.ProjectID,
			CreatedBy					:  venueType.CreatedBy,
			LastUpdateBy				:  venueType.LastUpdateBy,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetVenueTypeByCommercialTypeID(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUser(r)
    if !ok {
		view.RenderJSONError(w, "Failed get User for VenueTypes", http.StatusInternalServerError)
		return
    }
   _, ok = user["sub"]
   if !ok {
		c.reporter.Errorf("[handleGetAllVenueTypes] error get IDUser")
   }
	ctid, err := strconv.ParseInt(router.GetParam(r, "commercialTypeId"), 10, 64)

	if err != nil {
		c.reporter.Errorf("[handleGetVenueTypeByCommercialTypeID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	venueTypes, err := c.venueType.GetByCommercialType(10, ctid)
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
				Id							:  venueType.Id,
				Name						:  venueType.Name,
				Description					:  venueType.Description,
				Capacity					:  venueType.Capacity,
				CommercialTypeID			:  venueType.CommercialTypeID,
				PricingGroupID				:  venueType.PricingGroupID,
				CreatedAt					:  venueType.CreatedAt,
				UpdatedAt					:  venueType.UpdatedAt,
				DeletedAt					:  venueType.DeletedAt,
				Status						:  venueType.Status,
				ProjectID					:  venueType.ProjectID,
				CreatedBy					:  venueType.CreatedBy,
				LastUpdateBy				:  venueType.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

 