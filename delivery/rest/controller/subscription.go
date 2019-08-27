package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/subscription"

	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	var (
		pid = int64(10)
	)

	subscriptions, err := c.subscription.Select(pid)
	if err != nil {
		c.reporter.Errorf("[handleGetAllSubscriptions] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get Subscriptions", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		res = append(res, view.DataResponse{
			Type: "subscriptions",
			ID:   subscription.ID,
			Attributes: view.SubscriptionAttributes{
				PackageDuration: subscription.PackageDuration,
				BoxSerialNumber: subscription.BoxSerialNumber,
				SmartCardNumber: subscription.SmartCardNumber,
				OrderID:         subscription.OrderID,
				Status:          subscription.Status,
				ProjectID:       subscription.ProjectID,
				CreatedAt:       subscription.CreatedAt,
				UpdatedAt:       subscription.UpdatedAt,
				CreatedBy:       subscription.CreatedBy,
				LastUpdateBy:    subscription.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	var (
		pid     = int64(10)
		params  reqDeleteSubscription
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)
	if err != nil {
		c.reporter.Warningf("[handleDeleteSubscription] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.subscription.Get(pid, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteSubscription] subscription not found, err: %s", err.Error())
		view.RenderJSONError(w, "subscription not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteSubscription] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get subscription", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handleDeleteSubscription] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	//check user id
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handleDeleteSubscription] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	if !ok {
		_ = form.Bind(&params, r)
		if params.UserID == "" {
			c.reporter.Errorf("[handleDeleteSubscription] invalid parameter, failed get userID")
			view.RenderJSONError(w, "invalid parameter, failed get userID", http.StatusBadRequest)
			return
		}
		userID = params.UserID
		isAdmin = true
	}

	err = c.subscription.Delete(10, id, isAdmin, userID.(string))
	if err != nil {
		c.reporter.Errorf("[handleDeleteSubscription] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete subscription", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostSubscription(w http.ResponseWriter, r *http.Request) {
	var (
		params reqSubscription
		pid    = int64(10)
	)
	err := form.Bind(&params, r)

	//checking if userID nil, it will be request
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostSubscription] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	var uid = ""
	if !ok {
		uid = params.CreatedBy
	} else {
		uid = fmt.Sprintf("%v", userID)
	}

	if err != nil {
		c.reporter.Warningf("[handlePostSubscription] invalid parameter, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	subscription := subscription.Subscription{
		PackageDuration: params.PackageDuration,
		BoxSerialNumber: params.BoxSerialNumber,
		SmartCardNumber: params.SmartCardNumber,
		OrderID:         params.OrderID,
		ProjectID:       pid,
		CreatedBy:       uid,
	}

	err = c.subscription.Insert(&subscription)
	if err != nil {
		c.reporter.Infof("[handlePostSubscription] error insert subscription repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post subscription", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, subscription, http.StatusOK)
}

func (c *Controller) handlePatchSubscription(w http.ResponseWriter, r *http.Request) {
	var (
		pid     = int64(10)
		params  reqSubscription
		id, err = strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
		isAdmin = false
	)
	if err != nil {
		c.reporter.Warningf("[handlePatchSubscription] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchSubscription] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.subscription.Get(10, id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchSubscription] subscription not found, err: %s", err.Error())
		view.RenderJSONError(w, "Subscription not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchSubscription] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get subscription", http.StatusInternalServerError)
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

	subscription := subscription.Subscription{
		ID:              id,
		PackageDuration: params.PackageDuration,
		BoxSerialNumber: params.BoxSerialNumber,
		SmartCardNumber: params.SmartCardNumber,
		OrderID:         params.OrderID,
		ProjectID:       pid,
		LastUpdateBy:    userID.(string),
	}
	err = c.subscription.Update(&subscription, isAdmin)
	if err != nil {
		c.reporter.Errorf("[handlePatchSubscription] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update subscription", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, subscription, http.StatusOK)
}
