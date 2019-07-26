package controller

import (
	"database/sql"
	"fmt"
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
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
	_, isExist := c.admin.Check(fmt.Sprintf("%v", userID))
	if isExist == sql.ErrNoRows {
		c.reporter.Errorf("[handlePostEmailECert] user is not exist")
		view.RenderJSONError(w, "user is not exist", http.StatusUnauthorized)
		return
	}

	var params reqEmail

	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostEmailECert] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	/*content, sumorder, qrcodecontent := c.handleGetDataSertificate(params.OrderID, fmt.Sprintf("%s", userID))
	// content := c.handleGetDataInvoice(214, "kDQ2IAaHPZ8MTkqNS24zJPKu9MSLBo")
	htmlEmail := c.handleGetHtmlBodyCert(sumorder.VenueName)
	emailReq := email.EmailRequest{
		Subject: "Mola Live Arena E-Certificate",
		To:      sumorder.CompanyEmail,
		HTML:    htmlEmail,
		From:    "no-reply@molalivearena.com",
		Text:    " ",
		Attachments: []email.Attachment{
			{
				Content:     content,
				Filename:    "certificate.pdf",
				Type:        "plain/text",
				Disposition: "attachment",
				ContentID:   "contentid-test",
			},
			{
				Content:     qrcodecontent,
				Filename:    "molalivearena_qr.png",
				Type:        "image/png",
				Disposition: "attachment",
				ContentID:   "contentid-test",
			},
		},
	}
	errEmail := c.email.Send(emailReq)
	msg := c.handlePostEmailLog(userID, params.OrderID, emailReq.To, "ecert")
	if msg == "0" {
		c.reporter.Errorf("[handlePostEmailECert], err save email_log: %s", errEmail.Error())
	}
	if errEmail != nil {
		c.reporter.Errorf("[email failed to send], err: %s", errEmail.Error())
		view.RenderJSONData(w, false, http.StatusOK)
		return
	}*/

	view.RenderJSONData(w, true, http.StatusOK)
}
