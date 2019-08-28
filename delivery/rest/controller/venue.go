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
	var (
		venues venue.VenueAvailables
		err 	error
	)
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
	var (
		venues venue.VenueGroupAvailables
		err error
		projectID = int64(10)
	)
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

func (c *Controller) handleGetVenueByID(w http.ResponseWriter, r *http.Request) {
	var (
		_id     = router.GetParam(r, "id")
		id, err = strconv.ParseInt(_id, 10, 64)
		isAdmin = false
		userid  = ""
		venue 	venue.Venue
	)
	if err != nil {
		c.reporter.Errorf("[handleGetVenueByID] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetVenueByID] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}

	userID, ok := user["sub"]
	if !ok {
		isAdmin = true
	} else {
		userid = fmt.Sprintf("%v", userID)
	}

	if isAdmin == true {
		venue, err = c.venue.Get(10, id, "")
	} else {
		venue, err = c.venue.Get(10, id, userid)
	}
	if err != nil {
		c.reporter.Errorf("[handleGetVenueByID] venue not found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleGetVenueByID] failed get Venue, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get venue", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		Type: "venue",
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
	}
	view.RenderJSONData(w, res, http.StatusOK)
}


// Select All with Pagination
func (c *Controller) handleGetAllVenues(w http.ResponseWriter, r *http.Request) {
	var (
		getParam 		= r.URL.Query()
		cityName 		= getParam.Get("city")
		statusVenue 	= getParam.Get("status")
		limitVal 		= getParam.Get("limit")
		offsetVal 		= getParam.Get("page")
		showStatus 		= getParam.Get("show")
		projectID 		= int64(10)
		venues 			venue.Venues
		userid 			string
		err 			error
		limitBase		= 9
		offset 			= 1
		hasNext 		= false
	) 	
	
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
			userid = ""
		} else {
			userid = fmt.Sprintf("%v", userID)
		}
		venues, err = c.venue.Select(10, userid)
	}

	if err != nil {
		c.reporter.Errorf("[handleGetAllVenues] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Venues", http.StatusInternalServerError)
		return
	}

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

// Select All without Pagination
func (c *Controller) handleSelectAllVenues(w http.ResponseWriter, r *http.Request) {
	var (
		userid	string
		err 	error
		venues  venue.Venues
	)
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleGetAllVenues] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		userid = ""
	} else {
		userid = fmt.Sprintf("%v", userID)
	}
	venues, err = c.venue.Select(10, userid)

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

	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (c *Controller) handleDeleteVenue(w http.ResponseWriter, r *http.Request) {
	var (
		id, err 	= strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		userid 		string
		isAdmin 	= false
		params 		reqVenue
		venues 		venue.Venue
	)
	
	if err != nil {
		c.reporter.Warningf("[handleDeleteVenue] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteVenue] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	
	userID, ok := user["sub"]
	if !ok {
		err = form.Bind(&params, r)
		if err != nil {
			c.reporter.Warningf("[handleDeleteVenue] form binding, err: %s", err.Error())
			view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
			return
		}
		userid = params.CreatedBy
		isAdmin = true
	} else {
		userid = fmt.Sprintf("%v", userID)
	}

	if isAdmin == true {
		venues, err = c.venue.Get(10, id, "")
	} else {
		venues, err = c.venue.Get(10, id, userid)
	}

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

	err = c.venue.Delete(10, id, userid, venues.CreatedBy, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handleDeleteVenue] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete Venue", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostVenue(w http.ResponseWriter, r *http.Request) {
	var (
		params reqVenue
		userid string
	)
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostVenue] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostVenue] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	userID, ok := user["sub"]
	if !ok {
		userid = params.CreatedBy
	} else {
		userid = fmt.Sprintf("%v", userID)
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
		ShowStatus:                   params.ShowStatus,
		CreatedBy:                    userid,
	}

	err = c.venue.Insert(&venue)
	if err != nil {
		c.reporter.Infof("[handlePostVenue] error insert Venue repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post Venue", http.StatusInternalServerError)
		return
	}
	city, err := c.venue.GetCity(params.City)
	if len(city) == 0 {
		err = c.venue.InsertVenueAvailable(params.City, 1)
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
	var (
		id, err 	= strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		params 		reqVenue
		isAdmin 	= false
		userid 		string
		venues 		venue.Venue
	)
	
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

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchVenue] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	userID, ok := user["sub"]
	if !ok {
		userid = params.LastUpdateBy
		isAdmin = true
	} else {
		userid = fmt.Sprintf("%v", userID)
	}

	if isAdmin == true {
		venues, err = c.venue.Get(10, id, "")
	} else {
		venues, err = c.venue.Get(10, id, userid)
	}

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
		PtID:                         params.PtID,
		VenueCategory:                params.VenueCategory,
		PicName:                      params.PicName,
		PicContactNumber:             params.PicContactNumber,
		VenueTechnicianName:          params.VenueTechnicianName,
		VenueTechnicianContactNumber: params.VenueTechnicianContactNumber,
		VenuePhone:                   params.VenuePhone,
		ShowStatus:                   params.ShowStatus,
		CreatedBy:                    venues.CreatedBy,
		LastUpdateBy:                 userid,
	}
	err = c.venue.Update(&venue, userid, isAdmin)
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
			CreatedBy:        venue.CreatedBy,
			LastUpdateBy:     userid,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleShowStatusVenue(w http.ResponseWriter, r *http.Request) {
	var (
		id, err 	= strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin 	= false
		userid 		string
		status 		int64
		vaStatus 	int64
	)
	
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
		userid = ""
		isAdmin = true
	} else {
		userid = fmt.Sprintf("%v", userID)
	}

	venues, err := c.venue.GetStatus(10, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteVenue] Venue not found, err: %s", err.Error())
		view.RenderJSONError(w, "Venue not found", http.StatusNotFound)
		return
	}

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
		CreatedBy:        userid,
		LastUpdateBy:     userid,
		ShowStatus:       status,
	}
	err = c.venue.Update(&venue, userid, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePatchVenue] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update Venue", http.StatusInternalServerError)
		return
	}
	venuesAvailable, err := c.venue.GetCity(venues.City)

	for _, venuesAvailables := range venuesAvailable {
		vaStatus = venuesAvailables.Status
	}
	if status == 1 {
		if vaStatus == 0 {
			err = c.venue.UpdateStatusVenueAvailable(venues.City, 1)
		}
	} else {
		_, err := c.venue.GetVenueByCity(10, venues.City, "true", 9, 0)
		if err == sql.ErrNoRows {
			if vaStatus == 1 {
				err = c.venue.UpdateStatusVenueAvailable(venues.City, 0)
			}
		} else {
			err = c.venue.UpdateStatusVenueAvailable(venues.City, 1)
		}
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
			LastUpdateBy:     userid,
			ShowStatus:       status,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
	return
}

func (c *Controller) InsertLicense(venueID int64, createdBy string, buyerID string) error {
	var (
		licenseNumberUUID 	= util.GenerateUUID()
		layout 				= "2006-01-02T15:04:05.000Z"
		str 				= "1999-01-01T11:45:26.371Z"
		defaultTime, err 	= time.Parse(layout, str)
	)
	if err != nil {
		c.reporter.Warningf("[handleInsertLicense] Error insert license, err: %s", err.Error())
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
