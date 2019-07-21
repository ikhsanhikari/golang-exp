package controller

import (
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/email"
)

func (c *Controller) handlePostEmail(w http.ResponseWriter, r *http.Request) {
	var params reqEmail
	content := c.handleBasePdf(214, "kDQ2IAaHPZ8MTkqNS24zJPKu9MSLBo")
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
	}

	view.RenderJSONData(w, true, http.StatusOK)
}
