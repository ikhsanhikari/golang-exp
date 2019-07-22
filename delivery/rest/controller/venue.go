package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllVenuesAvailable(w http.ResponseWriter, r *http.Request) {
	var venues venue.VenueAvailables
	var err error
	venues, err = c.venue.GetVenueAvailable()

	if err != nil {
		c.reporter.Errorf("[handleGetAllVenues] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Venues", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(venues))
	for _, venue := range venues {
		res = append(res, view.DataResponse{
			Type: "city available",
			ID:   venue.Id,
			Attributes: view.VenueAvailableAttributes{
				Id:       venue.Id,
				CityName: venue.CityName,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllVenuesGroupAvailable(w http.ResponseWriter, r *http.Request) {
	var venues venue.VenueGroupAvailables
	var err error
	projectID := int64(10)
	venues, err = c.venue.GetVenueGroupAvailable(projectID)

	if err != nil {
		c.reporter.Errorf("[handleGetAllVenues] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Venues", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(venues))
	for _, venue := range venues {
		if venue.CityName != "" {
			res = append(res, view.DataResponse{
				Type: "Venue Group City",
				Attributes: view.VenueGroupAvailableAttributes{
					CityName: venue.CityName,
				},
			})
		}
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllVenues(w http.ResponseWriter, r *http.Request) {
	getParam := r.URL.Query()
	cityName := getParam.Get("city")
	statusVenue := getParam.Get("status")
	limitVal := getParam.Get("limit")
	offsetVal := getParam.Get("offset")
	projectID := int64(10)
	var venues venue.Venues
	var err error
	limit := 9
	offset := 0

	if limitVal != "" {
		limit, err = strconv.Atoi(limitVal)
	}
	if offsetVal != "" {
		offset, err = strconv.Atoi(offsetVal)
	}
	offset = limit * offset
	limit = limit + 1

	if cityName != "" && statusVenue != "true" {
		venues, err = c.venue.GetVenueByCity(projectID, cityName, limit, offset)
	} else if cityName == "all" && statusVenue == "true" {
		venues, err = c.venue.GetVenueByCity(projectID, cityName, limit, offset)
	} else if statusVenue == "true" && cityName == "" {
		venues, err = c.venue.GetVenueByStatus(projectID, limit, offset)
	} else if cityName != "all" && statusVenue == "true" {
		venues, err = c.venue.GetVenueByCityID(projectID, cityName, limit, offset)
	} else {
		user, ok := authpassport.GetUser(r)
		if !ok {
			c.reporter.Errorf("[handleGetAllVenues] failed get user")
			view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
			return
		}
		userID, ok := user["sub"]
		if !ok {
			c.reporter.Errorf("[handleGetAllVenues] failed get userID")
			view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
			return
		}
		venues, err = c.venue.Select(10, fmt.Sprintf("%v", userID))
	}

	if err != nil {
		c.reporter.Errorf("[handleGetAllVenues] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Venues", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(venues))
	for _, venue := range venues {
		res = append(res, view.DataResponse{
			Type: "venues",
			ID:   venue.Id,
			Attributes: view.VenueAttributes{
				Id:                           venue.Id,
				VenueId:                      venue.VenueId,
				VenueType:                    venue.VenueType,
				VenueName:                    venue.VenueName,
				Address:                      venue.Address,
				City:                         venue.City,
				Province:                     venue.Province,
				Zip:                          venue.Zip,
				Capacity:                     venue.Capacity,
				Facilities:                   venue.Facilities,
				PtID:                         venue.PtID,
				CreatedAt:                    venue.CreatedAt,
				UpdatedAt:                    venue.UpdatedAt,
				DeletedAt:                    venue.DeletedAt,
				Longitude:                    venue.Longitude,
				Latitude:                     venue.Latitude,
				Status:                       venue.Status,
				VenueCategory:                venue.VenueCategory,
				PicName:                      venue.PicName,
				PicContactNumber:             venue.PicContactNumber,
				VenueTechnicianName:          venue.VenueTechnicianName,
				VenueTechnicianContactNumber: venue.VenueTechnicianContactNumber,
				VenuePhone:                   venue.VenuePhone,
				CreatedBy:                    venue.CreatedBy,
				LastUpdateBy:                 venue.LastUpdateBy,
			},
		})
	}
	var hasNext bool
	if len(res) > 9 {
		hasNext = true
		view.RenderJSONDataPage(w, res, hasNext, http.StatusOK)
	} else {
		view.RenderJSONData(w, res, http.StatusOK)
	}

}

// Handle delete
func (c *Controller) handleDeleteVenue(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteVenue] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleDeleteVenue] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteVenue] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.venue.Get(10, id, fmt.Sprintf("%v", userID))
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteVenue] Venue not found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteVenue] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Venue", http.StatusInternalServerError)
		return
	}

	err = c.venue.Delete(10, id, fmt.Sprintf("%v", userID))
	if err != nil {
		c.reporter.Errorf("[handleDeleteVenue] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete Venue", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostVenue(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostVenue] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handlePostVenue] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	var params reqVenue
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostVenue] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	venue := venue.Venue{
		VenueId:                      params.VenueId,
		VenueType:                    params.VenueType,
		VenueName:                    params.VenueName,
		Address:                      params.Address,
		City:                         params.City,
		Province:                     params.Province,
		Zip:                          params.Zip,
		Capacity:                     params.Capacity,
		Facilities:                   params.Facilities,
		Longitude:                    params.Longitude,
		Latitude:                     params.Latitude,
		People:                       params.People,
		PtID:                         params.PtID,
		VenueCategory:                params.VenueCategory,
		PicName:                      params.PicName,
		PicContactNumber:             params.PicContactNumber,
		VenueTechnicianName:          params.VenueTechnicianName,
		VenueTechnicianContactNumber: params.VenueTechnicianContactNumber,
		VenuePhone:                   params.VenuePhone,
		CreatedBy:                    fmt.Sprintf("%v", userID),
	}

	err = c.venue.Insert(&venue)
	if err != nil {
		c.reporter.Infof("[handlePostVenue] error insert Venue repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post Venue", http.StatusInternalServerError)
		return
	}
	_, err = c.venue.GetCity(params.City)
	if err != sql.ErrNoRows {
		err = c.venue.InsertVenueAvailable(params.City)
	}

	view.RenderJSONData(w, venue, http.StatusOK)
}

func (c *Controller) handlePatchVenue(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchVenue] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePatchVenue] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handlePatchVenue] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	var params reqVenue
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchVenue] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.venue.Get(10, id, fmt.Sprintf("%v", userID))
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchVenue] Venue not found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchVenue] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Venue", http.StatusInternalServerError)
		return
	}

	venue := venue.Venue{
		Id:                           id,
		VenueId:                      params.VenueId,
		VenueType:                    params.VenueType,
		VenueName:                    params.VenueName,
		Address:                      params.Address,
		City:                         params.City,
		Province:                     params.Province,
		Zip:                          params.Zip,
		Capacity:                     params.Capacity,
		Facilities:                   params.Facilities,
		Longitude:                    params.Longitude,
		Latitude:                     params.Latitude,
		People:                       params.People,
		PtID:                         params.PtID,
		VenueCategory:                params.VenueCategory,
		PicName:                      params.PicName,
		PicContactNumber:             params.PicContactNumber,
		VenueTechnicianName:          params.VenueTechnicianName,
		VenueTechnicianContactNumber: params.VenueTechnicianContactNumber,
		VenuePhone:                   params.VenuePhone,
		LastUpdateBy:                 fmt.Sprintf("%v", userID),
	}
	err = c.venue.Update(&venue, fmt.Sprintf("%v", userID))
	if err != nil {
		c.reporter.Errorf("[handlePatchVenue] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update Venue", http.StatusInternalServerError)
		return
	}
	res := view.DataResponse{
		ID:   id,
		Type: "venues",
		Attributes: view.VenueAttributes{
			Id:                           id,
			VenueId:                      params.VenueId,
			VenueType:                    params.VenueType,
			VenueName:                    params.VenueName,
			Address:                      params.Address,
			City:                         params.City,
			Province:                     params.Province,
			Zip:                          params.Zip,
			Capacity:                     params.Capacity,
			Facilities:                   params.Facilities,
			Longitude:                    params.Longitude,
			Latitude:                     params.Latitude,
			People:                       params.People,
			PtID:                         params.PtID,
			UpdatedAt:                    time.Now(),
			Status:                       1,
			VenueCategory:                params.VenueCategory,
			PicName:                      params.PicName,
			PicContactNumber:             params.PicContactNumber,
			VenueTechnicianName:          params.VenueTechnicianName,
			VenueTechnicianContactNumber: params.VenueTechnicianContactNumber,
			VenuePhone:                   params.VenuePhone,
			LastUpdateBy:                 fmt.Sprintf("%v", userID),
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}
