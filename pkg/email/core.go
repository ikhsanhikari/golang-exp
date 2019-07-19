package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ICore is the interface
type ICore interface {
	Send(emailRequest EmailRequest) (err error)
}

// core contains db client
type core struct {
	apiBaseURL          string
	tokenGeneratorEmail TokenGeneratorEmail
}

var httpClient = http.Client{
	Timeout: time.Second * 10,
}

// this is the example to create http request
func (c *core) Send(emailRequest EmailRequest) (err error) {

	// accessToken, err := c.tokenGeneratorEmail.GetAccessToken(5)
	// if err != nil {
	// 	return err
	// }

	body, err := json.Marshal(EmailRequest{
		Subject: emailRequest.Subject,
		To:      emailRequest.To,
		HTML:    emailRequest.HTML,
		From:    emailRequest.From,
		Text:    emailRequest.Text,
	})

	// var url = c.apiBaseURL + "/v1/email/send"

	var url = "http://10.220.0.50/send"

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	// request.Header.Add("Authorization", "Bearer "+accessToken)

	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return err
	}
	fmt.Printf("Email Sent To : %s", emailRequest.To)
	return
}
