package email

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"net/http"
	"os"
	"time"

	"github.com/divan/qrlogo"
)

// ICore is the interface
type ICore interface {
	Send(emailRequest EmailRequest) (err error)
	GetBase64Png(licenseNum string) string
}

// core contains db client
type core struct {
	urlQrCode           string
	apiBaseURL          string
	tokenGeneratorEmail TokenGeneratorEmail
}

var httpClient = http.Client{
	Timeout: time.Second * 10,
}

// this is the example to create http request
func (c *core) Send(emailRequest EmailRequest) (err error) {
	// accessToken, err := c.tokenGeneratorEmail.GetAccessToken(5)
	if err != nil {
		return err
	}

	body, err := json.Marshal(EmailRequest{
		Subject:     emailRequest.Subject,
		To:          emailRequest.To,
		HTML:        emailRequest.HTML,
		From:        emailRequest.From,
		Text:        emailRequest.Text,
		Attachments: emailRequest.Attachments,
	})

	// var url = c.apiBaseURL + "/v1/email/send"
	var url = c.apiBaseURL + "/send"

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

var (
	input  = flag.String("i", "file/img/logo.png", "Logo to be placed over QR code")
	output = flag.String("o", "qr.png", "Output filename")
	size   = flag.Int("size", 512, "Image size in pixels")
)

func (c *core) GetBase64Png(licenseNum string) string {

	qrCode := c.urlQrCode + licenseNum

	fmt.Println(qrCode)

	file, err := os.Open(*input)
	errcheck(err, "Failed to open logo:")
	defer file.Close()

	logo, _, err := image.Decode(file)
	errcheck(err, "Failed to decode PNG with logo:")

	qr, err := qrlogo.Encode(qrCode, logo, *size)
	errcheck(err, "Failed to encode QR:")

	png := qr.Bytes()
	b64Png := base64.StdEncoding.EncodeToString(png)

	return b64Png
}

func errcheck(err error, str string) {
	if err != nil {
		fmt.Println(str, err)
		os.Exit(1)
	}
}
