package controller

import (
	"database/sql"
	"fmt"
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/email"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
)

func (c *Controller) handlePostEmailECert(w http.ResponseWriter, r *http.Request) {
	var (
		user, ok = authpassport.GetUser(r)
		params   reqEmail
	)

	if !ok {
		c.reporter.Errorf("[handlePostEmailECert] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusBadRequest)
		return
	}
	userID, ok := user["sub"].(string)
	if !ok {
		c.reporter.Errorf("[handlePostEmailECert] failed get userID")
		view.RenderJSONError(w, "failed get user", http.StatusBadRequest)
		return
	}

	//cek admin
	_, isExist := c.admin.Check(fmt.Sprintf("%v", userID))
	if isExist == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostEmailECert] user is not exist")
		view.RenderJSONError(w, "user is not exist", http.StatusUnauthorized)
		return
	}

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostEmailECert] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	result := c.handleEmailECert(params.VenueID, userID)

	if result == false {
		c.reporter.Warningf("[handlePostEmailECert] Error db or Email")
		view.RenderJSONError(w, "Error db or Email", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, true, http.StatusOK)
}

func (c *Controller) handleEmailECert(venueID int64, userID string) bool {
	content, sumvenue, qrcodecontent := c.handleGetDataSertificate(venueID, userID)

	if content == "0" {
		c.reporter.Errorf("[handleEmailECert] content get data invoice null")
		return false
	}

	htmlEmail := c.handleGetHtmlBodyCert(sumvenue.VenueName, sumvenue.VenueAddress)

	emailReq := email.EmailRequest{
		Subject: "Selamat! Keanggotaan Mola Live Arena sudah aktif.",
		To:      sumvenue.CompanyEmail,
		HTML:    htmlEmail,
		From:    "no-reply@molalivearena.com",
		Text:    " ",
		Attachments: []email.Attachment{
			{
				Content:     content,
				Filename:    "membership.pdf",
				Type:        "plain/text",
				Disposition: "attachment",
				ContentID:   "contentid-test",
			},
			{
				Content:     qrcodecontent,
				Filename:    "MolaLiveArena.png",
				Type:        "image/png",
				Disposition: "attachment",
				ContentID:   "contentid-test",
			},
		},
	}
	errEmail := c.email.Send(emailReq)
	msg := c.handlePostEmailEcertLog(userID, sumvenue.LastOrderID.Int64, venueID, emailReq.To, "ecert", sumvenue.CompanyID.Int64)

	if msg == "0" {
		c.reporter.Errorf("[handleEmailInvoice] Email Invoice Error")
		return false
	}
	if errEmail != nil {
		c.reporter.Errorf("[handleEmailInvoice] Error send Email Invoice ")
		return false
	}
	return true
}

func (c *Controller) handlePostEmailInvoice(w http.ResponseWriter, r *http.Request) {
	var (
		user, ok = authpassport.GetUser(r)
		params   reqInvoice
	)

	if !ok {
		c.reporter.Errorf("[handlePostEmailInvoice] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusBadRequest)
		return
	}
	userID, ok := user["sub"].(string)
	if !ok {
		c.reporter.Errorf("[handlePostEmailInvoice] failed get userID")
		view.RenderJSONError(w, "failed get user", http.StatusBadRequest)
		return
	}

	//cek admin
	_, isExist := c.admin.Check(fmt.Sprintf("%v", userID))
	if isExist == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostEmailInvoice] user is not exist")
		view.RenderJSONError(w, "user is not exist", http.StatusUnauthorized)
		return
	}

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostEmailInvoice] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	result := c.handleEmailInvoice(params.OrderID, fmt.Sprintf("%s", userID))
	if result == false {
		c.reporter.Warningf("[handlePostEmailInvoice] Error db or Email")
		view.RenderJSONError(w, "Error db or Email", http.StatusInternalServerError)
		return
	}
	view.RenderJSONData(w, true, http.StatusOK)
}

func (c *Controller) handleEmailInvoice(orderID int64, userID string) bool {
	var (
		em, name, address = "", "", ""
		venueID, compID   = int64(0), int64(0)
	)
	content, orderDetail := c.handleGetDataInvoice(orderID, userID)

	if content == "0" {
		c.reporter.Errorf("[handleEmailInvoice] content get data invoice null")
		return false
	}

	if len(orderDetail) > 0 {
		em = orderDetail[0].CompanyEmail
		venueID = orderDetail[0].VenueID
		compID = orderDetail[0].CompanyID
		name = orderDetail[0].VenueName
		address = orderDetail[0].Address
	} else {
		c.reporter.Errorf("[handleEmailInvoice] content get data invoice = 0")
		return false
	}

	htmlEmail := c.handleGetHtmlBodyInvoice(name, address)
	emailReq := email.EmailRequest{
		Subject: "Invoice",
		To:      em,
		HTML:    htmlEmail,
		From:    "no-reply@molalivearena.com",
		Text:    " ",
		Attachments: []email.Attachment{
			{
				Content:     content,
				Filename:    "invoice.pdf",
				Type:        "plain/text",
				Disposition: "attachment",
				ContentID:   "contentid-test",
			},
		},
	}
	errEmail := c.email.Send(emailReq)
	msg := c.handlePostEmailEcertLog(userID, orderID, venueID, emailReq.To, "invoice", compID)
	if msg == "0" {
		c.reporter.Errorf("[handleEmailInvoice] Email Invoice Error")
		return false
	}
	if errEmail != nil {
		c.reporter.Errorf("[handleEmailInvoice] Error send Email Invoice ")
		return false
	}
	return true
}
