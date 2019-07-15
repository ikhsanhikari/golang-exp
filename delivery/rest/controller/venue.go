package controller

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	//auth "git.sstv.io/lib/go/go-auth-api.git/authpassport"
)

func (c *Controller) handleGetAllVenues(w http.ResponseWriter, r *http.Request) {
	venues, err := c.venue.Select(10)
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
				Id							:  venue.Id,
				VenueId						:  venue.VenueId,
				VenueType					:  venue.VenueType,
				Address						:  venue.Address,
				Province					:  venue.Province,
				Zip							:  venue.Zip,
				Capacity					:  venue.Capacity,
				Facilities					:  venue.Facilities,
				CreatedAt					:  venue.CreatedAt,
				UpdatedAt					:  venue.UpdatedAt,
				DeletedAt					:  venue.DeletedAt,
				Longitude					:  venue.Longitude,
				Latitude					:  venue.Latitude,
				Status						:  venue.Status,
				VenueCategory				:  venue.VenueCategory,
				PicName						:  venue.PicName,
				PicContactNumber			:  venue.PicContactNumber,
				VenueTechnicianName			:  venue.VenueTechnicianName,
				VenueTechnicianContactNumber:  venue.VenueTechnicianContactNumber,
				VenuePhone					:  venue.VenuePhone,
				CreatedBy					:  venue.CreatedBy,
				LastUpdateBy				:  venue.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (c *Controller) handleDeleteVenue(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteVenue] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.venue.Get(10,id)
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

	err = c.venue.Delete(10,id)
	if err != nil {
		c.reporter.Errorf("[handleDeleteVenue] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete Venue", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostVenue(w http.ResponseWriter, r *http.Request) {
	var params reqVenue
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostVenue] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	
	venue := venue.Venue{
		VenueId						:  params.VenueId,
		VenueType					:  params.VenueType,
		Address						:  params.Address,
		Province					:  params.Province,
		Zip							:  params.Zip,
		Capacity					:  params.Capacity,
		Facilities					:  params.Facilities,
		Longitude					:  params.Longitude,
		Latitude					:  params.Latitude,
		People						:  params.People,
		VenueCategory				:  params.VenueCategory,
		PicName						:  params.PicName,
		PicContactNumber			:  params.PicContactNumber,
		VenueTechnicianName			:  params.VenueTechnicianName,
		VenueTechnicianContactNumber:  params.VenueTechnicianContactNumber,
		VenuePhone					:  params.VenuePhone,
		CreatedBy					:  params.CreatedBy,
	}

	err = c.venue.Insert(&venue)
	if err != nil {
		c.reporter.Infof("[handlePostVenue] error insert Venue repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post Venue", http.StatusInternalServerError)
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

	var params reqVenue
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchVenue] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.venue.Get(10,id)
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
		Id							: 	id,
		VenueId						:  params.VenueId,
		VenueType					:  params.VenueType,
		Address						:  params.Address,
		Province					:  params.Province,
		Zip							:  params.Zip,
		Capacity					:  params.Capacity,
		Facilities					:  params.Facilities,
		Longitude					:  params.Longitude,
		Latitude					:  params.Latitude,
		People						:  params.People,
		VenueCategory				:  params.VenueCategory,
		PicName						:  params.PicName,
		PicContactNumber			:  params.PicContactNumber,
		VenueTechnicianName			:  params.VenueTechnicianName,
		VenueTechnicianContactNumber:  params.VenueTechnicianContactNumber,
		VenuePhone					:  params.VenuePhone,
		LastUpdateBy				:  params.LastUpdateBy,
	}
	err = c.venue.Update(&venue)
	if err != nil {
		c.reporter.Errorf("[handlePatchVenue] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update Venue", http.StatusInternalServerError)
		return
	}
	res := view.DataResponse{
		ID:   id,
		Type: "venues",
		Attributes: view.VenueAttributes{
			Id								: 	id,
			VenueId							:   params.VenueId,
			VenueType						:  	params.VenueType,
			Address							:  	params.Address,
			Province						:  	params.Province,
			Zip								:   params.Zip,
			Capacity						:   params.Capacity,
			Facilities						:   params.Facilities,
			Longitude						:   params.Longitude,
			Latitude						:   params.Latitude,
			People							:   params.People,
			UpdatedAt						:   time.Now(),
			Status							:   1,
			VenueCategory					:   params.VenueCategory,
			PicName							:   params.PicName,
			PicContactNumber				:	params.PicContactNumber,
			VenueTechnicianName				:	params.VenueTechnicianName,
			VenueTechnicianContactNumber	:	params.VenueTechnicianContactNumber,
			VenuePhone						:  params.VenuePhone,
			LastUpdateBy					:  params.LastUpdateBy,
		},
	}

	view.RenderJSONData(w, res, http.StatusOK)
}
 