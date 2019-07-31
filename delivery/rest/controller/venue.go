package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/license"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	"git.sstv.io/lib/go/gojunkyard.git/util"
	null "gopkg.in/guregu/null.v3"
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
	offsetVal := getParam.Get("page")
	showStatus := getParam.Get("show")
	projectID := int64(10)
	var venues venue.Venues
	var err error
	limitBase := 9
	offset := 1
	if showStatus != "false" {
		showStatus = "true"
	} else {
		showStatus = "false"
	}
	if limitVal != "" {
		limitBase, err = strconv.Atoi(limitVal)
	}
	if offsetVal != "" {
		offset, err = strconv.Atoi(offsetVal)
	}
	offset = offset - 1
	offset = limitBase * offset
	limit := limitBase + 1

	if cityName != "" && statusVenue != "true" {
		//get Venue with cityName
		venues, err = c.venue.GetVenueByCity(projectID, cityName, showStatus, limit, offset)
	} else if statusVenue == "true" && cityName == "" {
		//get All Venue with status 2 /4
		venues, err = c.venue.GetVenueByStatus(projectID, limit, offset)
	} else if cityName != "all" && statusVenue == "true" {
		//get Venue with cityNMe & status 2 /4
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

	var hasNext bool
	hasNext = false
	res := make([]view.DataResponse, 0, len(venues))
	for num, venue := range venues {
		if num < limitBase {
			res = append(res, view.DataResponse{
				Type: "venues",
				ID:   venue.Id,
				Attributes: view.VenueAttributes{
					Id:               venue.Id,
					VenueId:          venue.VenueId,
					VenueType:        venue.VenueType,
					VenueName:        venue.VenueName,
					Address:          venue.Address,
					City:             venue.City,
					Province:         venue.Province,
					Zip:              venue.Zip,
					Capacity:         venue.Capacity,
					Facilities:       venue.Facilities,
					PtID:             venue.PtID,
					CreatedAt:        venue.CreatedAt,
					UpdatedAt:        venue.UpdatedAt,
					DeletedAt:        venue.DeletedAt,
					Longitude:        venue.Longitude,
					Latitude:         venue.Latitude,
					Status:           venue.Status,
					PicName:          venue.PicName,
					PicContactNumber: venue.PicContactNumber,
					VenuePhone:       venue.VenuePhone,
					CreatedBy:        venue.CreatedBy,
					LastUpdateBy:     venue.LastUpdateBy,
					ShowStatus:       venue.ShowStatus,
				},
			})
		}
		if num >= limitBase {
			hasNext = true
		}
	}
	view.RenderJSONDataPage(w, res, hasNext, http.StatusOK)

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
		People:                       null.IntFrom(params.People),
		PtID:                         params.PtID,
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
	if err == sql.ErrNoRows {
		err = c.venue.InsertVenueAvailable(params.City)
	}

	err = c.InsertLicense(venue.Id, venue.CreatedBy, venue.CreatedBy)
	if err != nil {
		c.reporter.Infof("[handlePostVenue] Failed post license, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post license", http.StatusInternalServerError)
		return
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
		People:                       null.IntFrom(params.People),
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
			Id:               id,
			VenueId:          params.VenueId,
			VenueType:        params.VenueType,
			VenueName:        params.VenueName,
			Address:          params.Address,
			City:             params.City,
			Province:         params.Province,
			Zip:              params.Zip,
			Capacity:         params.Capacity,
			Facilities:       params.Facilities,
			Longitude:        params.Longitude,
			Latitude:         params.Latitude,
			PtID:             params.PtID,
			UpdatedAt:        time.Now(),
			Status:           1,
			ShowStatus:       params.ShowStatus,
			PicName:          params.PicName,
			PicContactNumber: params.PicContactNumber,
			VenuePhone:       params.VenuePhone,
			LastUpdateBy:     fmt.Sprintf("%v", userID),
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleShowStatusVenue(w http.ResponseWriter, r *http.Request) {
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
	venues, err := c.venue.GetStatus(10, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteVenue] Venue not found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue not found", http.StatusNotFound)
		return
	}
	var status int64
	if venues.ShowStatus == 1 {
		status = 0
	} else {
		status = 1
	}
	venue := venue.Venue{
		Id:               id,
		VenueId:          venues.VenueId,
		VenueType:        venues.VenueType,
		VenueName:        venues.VenueName,
		Address:          venues.Address,
		City:             venues.City,
		Province:         venues.Province,
		Zip:              venues.Zip,
		Capacity:         venues.Capacity,
		Facilities:       venues.Facilities,
		Longitude:        venues.Longitude,
		Latitude:         venues.Latitude,
		PtID:             venues.PtID,
		UpdatedAt:        time.Now(),
		Status:           venues.Status,
		PicName:          venues.PicName,
		PicContactNumber: venues.PicContactNumber,
		VenuePhone:       venues.VenuePhone,
		LastUpdateBy:     fmt.Sprintf("%v", userID),
		ShowStatus:       status,
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
			Id:               id,
			VenueId:          venues.VenueId,
			VenueType:        venues.VenueType,
			VenueName:        venues.VenueName,
			Address:          venues.Address,
			City:             venues.City,
			Province:         venues.Province,
			Zip:              venues.Zip,
			Capacity:         venues.Capacity,
			Facilities:       venues.Facilities,
			Longitude:        venues.Longitude,
			Latitude:         venues.Latitude,
			PtID:             venues.PtID,
			UpdatedAt:        time.Now(),
			Status:           venues.Status,
			PicName:          venues.PicName,
			PicContactNumber: venues.PicContactNumber,
			VenuePhone:       venues.VenuePhone,
			LastUpdateBy:     fmt.Sprintf("%v", userID),
			ShowStatus:       status,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
	return
}

func (c *Controller) InsertLicense(venueID int64, createdBy string, buyerID string) error {
	licenseNumberUUID := util.GenerateUUID()
	layout := "2006-01-02T15:04:05.000Z"
	str := "1999-01-01T11:45:26.371Z"
	defaultTime, err := time.Parse(layout, str)
	if err != nil {
		fmt.Println(err)
	}
	license := license.License{
		LicenseNumber: licenseNumberUUID,
		OrderID:       venueID,
		LicenseStatus: 1,
		ActiveDate:    defaultTime,
		ExpiredDate:   defaultTime,
		ProjectID:     10,
		CreatedBy:     createdBy,
		BuyerID:       buyerID,
	}

	err = c.license.Insert(&license)
	if err != nil {
		return err
	}

	return nil
}
