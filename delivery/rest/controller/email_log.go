package controller

import (

	// "git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/email_log"
	// "git.sstv.io/lib/go/go-auth-api.git/authpassport"
	// "git.sstv.io/lib/go/gojunkyard.git/form"
	// "git.sstv.io/lib/go/gojunkyard.git/router"
	//auth "git.sstv.io/lib/go/go-auth-api.git/authpassport"
)

func (c *Controller) handlePostEmailEcertLog(userID string, orderID int64, venueID int64, to string, emailType string,compId int64) string {

	//companyEmail, err = c.company.GetByOrderID(orderID, 10)

	emailLog := email_log.EmailLog{
		SenderUID: userID,
		OrderID:   orderID,
		VenueID:   venueID,
		To:        to, //companyEmail.CompanyEmail,
		EmailType: emailType,
		CreatedBy: userID,
		CompanyID: compId,
	}

	err := c.emailLog.Insert(&emailLog)
	if err != nil {
		c.reporter.Infof("[handlePostEmailLog] error insert EmailLog repository, err: %s", err.Error())
		return "0"
	}
	return "1"

}
