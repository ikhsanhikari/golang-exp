package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/agent"
)

func (c *Controller) handleGetAllAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := c.agent.Select(10)
	if err != nil {
		c.reporter.Errorf("[handleGetAllAgents] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Agents", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(agents))
	for _, agent := range agents {
		res = append(res, view.DataResponse{
			Type: "agents",
			ID:   agent.ID,
			Attributes: view.AgentAttributes{
				UserID:       agent.UserID,
				Status:       agent.Status,
				ProjectID:    agent.ProjectID,
				CreatedAt:    agent.CreatedAt,
				UpdatedAt:    agent.UpdatedAt,
				CreatedBy:    agent.CreatedBy,
				LastUpdateBy: agent.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllAgentsByUserID(w http.ResponseWriter, r *http.Request) {
	id := router.GetParam(r, "userId")

	agents, err := c.agent.SelectByUserID(10, id)
	if err != nil {
		c.reporter.Errorf("[handleGetAllAgentsByUserID] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Agents", http.StatusInternalServerError)
		return
	}
	status := false
	if len(agents) > 0 {
		status = true
		view.RenderJSONData(w, status, http.StatusOK)
		return
	}
	view.RenderJSONData(w, status, http.StatusOK)
}

func (c *Controller) handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteAgent] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	agentParam, err := c.agent.Get(10, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteAgent] agent not found, err: %s", err.Error())
		view.RenderJSONError(w, "agent not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteAgent] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get agent", http.StatusInternalServerError)
		return
	}

	err = c.agent.Delete(10, id, agentParam.UserID)
	if err != nil {
		c.reporter.Errorf("[handleDeleteAgent] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete agent", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostAgent(w http.ResponseWriter, r *http.Request) {
	var params reqAgent
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostAgent] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	agent := agent.Agent{
		UserID:    params.UserID,
		ProjectID: 10,
		CreatedBy: params.CreatedBy,
	}

	err = c.agent.Insert(&agent)
	if err != nil {
		c.reporter.Infof("[handlePostAgent] error insert agent repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post agent", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, agent, http.StatusOK)
}

func (c *Controller) handlePatchAgent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchAgent] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqAgent
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchAgent] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	agentParam, err := c.agent.Get(10, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchAgent] agent not found, err: %s", err.Error())
		view.RenderJSONError(w, "Agent not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchAgent] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get agent", http.StatusInternalServerError)
		return
	}

	agent := agent.Agent{
		ID:           id,
		UserID:       params.UserID,
		ProjectID:    10,
		LastUpdateBy: params.LastUpdateBy,
	}
	err = c.agent.Update(&agent, agentParam.UserID)
	if err != nil {
		c.reporter.Errorf("[handlePatchAgent] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update agent", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, agent, http.StatusOK)
}

func (c *Controller) handleAgentsCheck(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleAgentsCheck] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}

	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handleAgentsCheck] failed get userID")
		view.RenderJSONError(w, "failed get userID", http.StatusInternalServerError)
		return
	}

	_, isExist := c.agent.Check(fmt.Sprintf("%v", userID))
	if isExist == sql.ErrNoRows {
		c.reporter.Errorf("[handleAgentsCheck] user is not exist")
		view.RenderJSONError(w, "user is not exist", http.StatusUnauthorized)
		return
	}

	view.RenderJSON(w, nil, http.StatusOK)
}
