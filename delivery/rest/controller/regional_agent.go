package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	regionalAgent "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/regional_agent"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllRegionalAgents(w http.ResponseWriter, r *http.Request) {
	var (
		pid = c.projectID
	)
	regionalAgents, err := c.regionalAgent.Select(pid)
	if err != nil {
		c.reporter.Errorf("[handleGetAllRegionalAgents] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get RegionalAgents", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(regionalAgents))
	for _, regionalAgent := range regionalAgents {
		res = append(res, view.DataResponse{
			Type: "regionalAgents",
			ID:   regionalAgent.ID,
			Attributes: view.RegionalAgentAttributes{
				Name:         regionalAgent.Name,
				Area:         regionalAgent.Area,
				Email:        regionalAgent.Email,
				Phone:        regionalAgent.Phone,
				Website:      regionalAgent.Website,
				Status:       regionalAgent.Status,
				ProjectID:    regionalAgent.ProjectID,
				CreatedAt:    regionalAgent.CreatedAt,
				UpdatedAt:    regionalAgent.UpdatedAt,
				CreatedBy:    regionalAgent.CreatedBy,
				LastUpdateBy: regionalAgent.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetRegionalAgents(w http.ResponseWriter, r *http.Request) {
	var (
		pid     = c.projectID
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	)
	regionalAgent, err := c.regionalAgent.Get(pid, id)
	if err != nil {
		c.reporter.Errorf("[handleGetRegionalAgents] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get RegionalAgents", http.StatusInternalServerError)
		return
	}

	res := view.DataResponse{
		Type: "regionalAgents",
		ID:   regionalAgent.ID,
		Attributes: view.RegionalAgentAttributes{
			Name:         regionalAgent.Name,
			Area:         regionalAgent.Area,
			Email:        regionalAgent.Email,
			Phone:        regionalAgent.Phone,
			Website:      regionalAgent.Website,
			Status:       regionalAgent.Status,
			ProjectID:    regionalAgent.ProjectID,
			CreatedAt:    regionalAgent.CreatedAt,
			UpdatedAt:    regionalAgent.UpdatedAt,
			CreatedBy:    regionalAgent.CreatedBy,
			LastUpdateBy: regionalAgent.LastUpdateBy,
		},
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteRegionalAgent(w http.ResponseWriter, r *http.Request) {
	var (
		pid     = c.projectID
		params  reqDeleteRegionalAgent
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)

	if err != nil {
		c.reporter.Warningf("[handleDeleteRegionalAgent] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.regionalAgent.Get(pid, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteRegionalAgent] regionalAgent not found, err: %s", err.Error())
		view.RenderJSONError(w, "regionalAgent not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteRegionalAgent] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get regionalAgent", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handleDeleteRegionalAgent] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	//check user id
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteRegionalAgent] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handleDeleteRegionalAgent] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}
	err = c.regionalAgent.Delete(pid, id, isAdmin, userID.(string))
	if err != nil {
		c.reporter.Errorf("[handleDeleteRegionalAgent] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete regionalAgent", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostRegionalAgent(w http.ResponseWriter, r *http.Request) {
	var (
		params reqRegionalAgent
		pid    = c.projectID
	)
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostRegionalAgent] id must be integer, err: %s", err.Error())
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

	regionalAgent := regionalAgent.RegionalAgent{
		Name:      params.Name,
		Area:      params.Area,
		Email:     params.Email,
		Phone:     params.Phone,
		Website:   params.Website,
		ProjectID: pid,
		CreatedBy: uid,
	}

	err = c.regionalAgent.Insert(&regionalAgent)
	if err != nil {
		c.reporter.Infof("[handlePostRegionalAgent] error insert regionalAgent repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post regionalAgent", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, regionalAgent, http.StatusOK)
}

func (c *Controller) handlePatchRegionalAgent(w http.ResponseWriter, r *http.Request) {
	var (
		pid     = c.projectID
		params  reqRegionalAgent
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)

	if err != nil {
		c.reporter.Warningf("[handlePatchRegionalAgent] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchRegionalAgent] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.regionalAgent.Get(pid, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchRegionalAgent] regionalAgent not found, err: %s", err.Error())
		view.RenderJSONError(w, "RegionalAgent not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchRegionalAgent] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get regionalAgent", http.StatusInternalServerError)
		return
	}

	//check user id
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePatchRegionalAgent] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.LastUpdateBy == "" {
			c.reporter.Errorf("[handlePatchRegionalAgent] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.LastUpdateBy
		isAdmin = true
	}

	regionalAgent := regionalAgent.RegionalAgent{
		ID:           id,
		Name:         params.Name,
		Area:         params.Area,
		Email:        params.Email,
		Phone:        params.Phone,
		Website:      params.Website,
		ProjectID:    pid,
		LastUpdateBy: userID.(string),
	}
	err = c.regionalAgent.Update(&regionalAgent, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePatchRegionalAgent] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update regionalAgent", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, regionalAgent, http.StatusOK)
}
