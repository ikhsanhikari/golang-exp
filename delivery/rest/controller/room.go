package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/room"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllRooms(w http.ResponseWriter, r *http.Request) {
	var (
		pid = c.projectID
	)
	rooms, err := c.room.Select(pid)
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
				Name:         room.Name,
				Description:  room.Description,
				Price:        room.Price,
				Status:       room.Status,
				ProjectID:    room.ProjectID,
				CreatedAt:    room.CreatedAt,
				UpdatedAt:    room.UpdatedAt,
				CreatedBy:    room.CreatedBy,
				LastUpdateBy: room.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	var (
		pid     = c.projectID
		params  reqDeleteRoom
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)

	if err != nil {
		c.reporter.Warningf("[handleDeleteRoom] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.room.Get(pid, id)
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

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handleDeleteRoom] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	//check user id
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteRoom] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handleDeleteRoom] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}
	err = c.room.Delete(pid, id, isAdmin, userID.(string))
	if err != nil {
		c.reporter.Errorf("[handleDeleteRoom] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete room", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostRoom(w http.ResponseWriter, r *http.Request) {
	var (
		params reqRoom
		pid    = c.projectID
	)
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostRoom] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	//checking if userID nil, it will be request
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostDevice] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	var uid = ""
	if !ok {
		//is Admin
		uid = params.CreatedBy
	} else {
		//is User
		uid = fmt.Sprintf("%v", userID)
	}

	room := room.Room{
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		ProjectID:   pid,
		CreatedBy:   uid,
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
	var (
		pid     = c.projectID
		params  reqRoom
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)

	if err != nil {
		c.reporter.Warningf("[handlePatchRoom] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchRoom] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.room.Get(pid, id)
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

	//check user id
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePatchRoom] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.LastUpdateBy == "" {
			c.reporter.Errorf("[handlePatchRoom] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.LastUpdateBy
		isAdmin = true
	}

	room := room.Room{
		ID:           id,
		Name:         params.Name,
		Description:  params.Description,
		Price:        params.Price,
		ProjectID:    pid,
		LastUpdateBy: userID.(string),
	}
	err = c.room.Update(&room, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePatchRoom] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update room", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, room, http.StatusOK)
}
