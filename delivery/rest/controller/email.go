package controller

import (
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
	userID, ok := user["sub"]
	if !ok {
		c.reporter.Errorf("[handlePostEmailECert] failed get userID")
		view.RenderJSONError(w, "failed get user", http.StatusBadRequest)
		return
	}

	// lakukan pengecekan userid harus admin

	var params reqEmail
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostEmailECert] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	content, sumorder := c.handleGetDataSertificate(params.OrderID, fmt.Sprintf("%s", userID))
	// content := c.handleGetDataInvoice(214, "kDQ2IAaHPZ8MTkqNS24zJPKu9MSLBo")
	emailReq := email.EmailRequest{
		Subject: "Mola Live Arena E-Certificate",
		To:      sumorder.CompanyEmail,
		HTML:    "<h1>ISI PESAN !!!!!!!</h1>",
		From:    "no-reply@molalivearena.com",
		Text:    "...",
		Attachments: []email.Attachment{
			{
				Content:     content,
				Filename:    "certificate.pdf",
				Type:        "plain/text",
				Disposition: "attachment",
				ContentID:   "contentid-test",
			},
		},
	}
	errEmail := c.email.Send(emailReq)
	msg := c.handlePostEmailLog("kDQ2IAaHPZ8MTkqNS24zJPKu9MSLBo", params.OrderID, emailReq.To, "ecert")
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
