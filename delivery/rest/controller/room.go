package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/room"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := c.room.Select(10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllRooms] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Rooms", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(rooms))
	for _, room := range rooms {
		res = append(res, view.DataResponse{
			Type: "rooms",
			ID:   room.ID,
			Attributes: view.RoomAttributes{
				Name:        room.Name,
				Description: room.Description,
				Price:       room.Price,
				Status:      room.Status,
				ProjectID:   room.ProjectID,
				CreatedAt:   room.CreatedAt,
				UpdatedAt:   room.UpdatedAt,
				CreatedBy:		room.CreatedBy,
				LastUpdateBy: room.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteRoom] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.room.Get(10, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteRoom] room not found, err: %s", err.Error())
		view.RenderJSONError(w, "room not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteRoom] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get room", http.StatusInternalServerError)
		return
	}

	err = c.room.Delete(10, id)
	if err != nil {
		c.reporter.Errorf("[handleDeleteRoom] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete room", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostRoom(w http.ResponseWriter, r *http.Request) {
	var params reqRoom
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostRoom] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	room := room.Room{
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		ProjectID:   10,
		CreatedBy: 		params.CreatedBy,
	}

	err = c.room.Insert(&room)
	if err != nil {
		c.reporter.Infof("[handlePostRoom] error insert room repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post room", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, room, http.StatusOK)
}

func (c *Controller) handlePatchRoom(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchRoom] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqRoom
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchRoom] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.room.Get(10, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchRoom] room not found, err: %s", err.Error())
		view.RenderJSONError(w, "Room not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchRoom] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get room", http.StatusInternalServerError)
		return
	}

	room := room.Room{
		ID:          id,
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		ProjectID:   10,
		LastUpdateBy: params.LastUpdateBy,
	}
	err = c.room.Update(&room)
	if err != nil {
		c.reporter.Errorf("[handlePatchRoom] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update room", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, room, http.StatusOK)
}
