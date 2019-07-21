package controller

import (
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/email"
	"git.sstv.io/lib/go/gojunkyard.git/form"
)

func (c *Controller) handlePostEmail(w http.ResponseWriter, r *http.Request) {
	var params reqEmail
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostDevice] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}
	content := c.handleBasePdf(params.OrderID, "kDQ2IAaHPZ8MTkqNS24zJPKu9MSLBo")
	emailReq := email.EmailRequest{
		Subject: "subject is nothing !!!",
		To:      params.To,
		HTML:    "<h1>ISI PESAN !!!!!!!</h1>",
		From:    "no-reply@molalivearena.com",
		Text:    "Empty Text ........",
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
	if errEmail != nil {
		c.reporter.Errorf("[email failed to send], err: %s", errEmail.Error())
		view.RenderJSONData(w, false, http.StatusOK)
		return
	}

	view.RenderJSONData(w, true, http.StatusOK)
}
