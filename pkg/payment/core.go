package payment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ICore is the interface
type ICore interface {
	Pay(id string, paymentMethodID int64) (payment *Payment, err error)
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
func (c *core) Pay(id string, paymentMethodID int64) (payment *Payment, err error) {
	accessToken, err := c.tokenGenerator.GetAccessToken(10)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]interface{}{
		"payment_method_id": paymentMethodID,
		"id":                id,
	})

	var url = c.apiBaseURL + "/api/v2/payments/api/v1/dopay_molanobar?app_id=molalivearena"

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("error ini:", err)
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+accessToken)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, err
	}

	err = json.NewDecoder(response.Body).Decode(&payment)
	if err != nil {
		return nil, err
	}

	return payment, nil
}
