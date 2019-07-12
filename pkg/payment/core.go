package payment

import (
	"encoding/json"
	"net/http"
	"time"
)

// ICore is the interface
type ICore interface {
	Pay() error
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
func (c *core) Pay() error {
	accessToken, err := c.tokenGenerator.GetAccessToken(2)
	if err != nil {
		return err
	}

	var url = c.apiBaseURL + "/path"
	request, err := http.NewRequest("", url, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)

	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(nil)
	if err != nil {
		return err
	}

	// do something

	return nil
}
