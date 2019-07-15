package payment

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ICore is the interface
type ICore interface {
	Pay(id string, paymentMethodID int64) (payment Payment, err error)
}

// core contains db client
type core struct {
	apiBaseURL     string
	tokenGenerator TokenGenerator
}

var httpClient = http.Client{
	Timeout: time.Second * 10,
}

// this is the example to create http request
func (c *core) Pay(id string, paymentMethodID int64) (payment Payment, err error) {
	log.Println(id, paymentMethodID)

	accessToken, err := c.tokenGenerator.GetAccessToken(10)
	if err != nil {
		return payment, err
	}
	log.Println(accessToken)

	var param = url.Values{}
	param.Set("payment_method_id", strconv.FormatInt(paymentMethodID, 10))
	param.Set("id", id)
	var payload = bytes.NewBufferString(param.Encode())

	var url = c.apiBaseURL + "/api/v2/payments/api/v1/dopay_molanobar?app_id=molalivearena"

	request, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return payment, err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+accessToken)

	response, err := httpClient.Do(request)
	log.Println(response)
	if err != nil {
		return payment, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return payment, err
	}

	err = json.NewDecoder(response.Body).Decode(&payment)
	if err != nil {
		return payment, err
	}

	return payment, nil
}
