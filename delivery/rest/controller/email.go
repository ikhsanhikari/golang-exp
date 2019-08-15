package controller

import (
	//"database/sql"
	"fmt"
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/email"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
)

func (c *Controller) handlePostEmailECert(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
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

	// cek admin
	// _, isExist := c.admin.Check(fmt.Sprintf("%v", userID))
	// if isExist == sql.ErrNoRows {
	// 	c.reporter.Errorf("[handlePostEmailECert] user is not exist")
	// 	view.RenderJSONError(w, "user is not exist", http.StatusUnauthorized)
	// 	return
	// }

	var params reqEmail

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostEmailECert] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	content, sumvenue, qrcodecontent := c.handleGetDataSertificate(params.VenueID, fmt.Sprintf("%s", userID))
	htmlEmail := c.handleGetHtmlBodyCert(sumvenue.VenueName)
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
	msg := c.handlePostEmailEcertLog(userID, sumvenue.LastOrderID, params.VenueID, emailReq.To, "ecert", sumvenue.CompanyID)
	if msg == "0" {
		c.reporter.Errorf("[handlePostEmailECert], err save email_log: %s", errEmail.Error())
	}
	if errEmail != nil {
		c.reporter.Errorf("[email failed to send], err: %s", errEmail.Error())
		view.RenderJSONData(w, false, http.StatusOK)
		return
	}

	view.RenderJSONData(w, true, http.StatusOK)
}

func (c *Controller) handlePostEmailInvoice(w http.ResponseWriter, r *http.Request) {
	user, ok := authpassport.GetUser(r)
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

	// cek admin
	// _, isExist := c.admin.Check(fmt.Sprintf("%v", userID))
	// if isExist == sql.ErrNoRows {
	// 	c.reporter.Errorf("[handlePostEmailInvoice] user is not exist")
	// 	view.RenderJSONError(w, "user is not exist", http.StatusUnauthorized)
	// 	return
	// }

	var params reqInvoice

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostEmailInvoice] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	content, orderDetail := c.handleGetDataInvoice(params.OrderID, fmt.Sprintf("%s", userID))

	em,venueID,compID := "", int64(0), int64(0)
	if len(orderDetail) > 0 {
		em = orderDetail[0].CompanyEmail
		venueID = orderDetail[0].VenueID
		compID = orderDetail[0].CompanyID
	}

	htmlEmail := c.handleGetHtmlBodyCert("a")
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
	msg := c.handlePostEmailEcertLog(userID, params.OrderID, venueID, emailReq.To, "invoice", compID)
	if msg == "0" {
		c.reporter.Errorf("[handlePostEmailInvoice], err save email_log: %s", errEmail.Error())
	}
	if errEmail != nil {
		c.reporter.Errorf("[email failed to send], err: %s", errEmail.Error())
		view.RenderJSONData(w, false, http.StatusOK)
		return
	}

	view.RenderJSONData(w, true, http.StatusOK)
}
