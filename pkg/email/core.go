package email

import (
	"bufio"
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
	GetBase64Png(licenseNum string) (string, string)
	GetPic() (string)
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
	input  = flag.String("i", "file/img/mola.png", "Logo to be placed over QR code")
	output = flag.String("o", "qr.png", "Output filename")
	size   = flag.Int("size", 340, "Image size in pixels")
)

func (c *core) GetBase64Png(licenseNum string) (string, string) {

	qrCode := c.urlQrCode + licenseNum

	file, err := os.Open(*input)
	if err != nil {
		return "0", "0"
	}
	defer file.Close()

	logo, _, err := image.Decode(file)
	if err != nil {
		return "0", "0"
	}
	qr, err := qrlogo.Encode(qrCode, logo, *size)
	if err != nil {
		return "0", "0"
	}

	png := qr.Bytes()
	b64Png := base64.StdEncoding.EncodeToString(png)

	pathfile := "file/img/background1.png"

	b64Pd := GetPng(pathfile)

	return b64Png, b64Pd
}

func (c *core) GetPic() (string) {

	pathfile := "file/img/mix.png"

	b64 := GetPng(pathfile)

	return b64
}

func GetPng(pic string) (string) {

	file, err := os.Open(pic)

	if err != nil {
		return "0"
	}

	fInfo, _ := file.Stat()
	var size int64 = fInfo.Size()
	buf := make([]byte, size)

	fReader := bufio.NewReader(file)
	fReader.Read(buf)

	b64Pd := base64.StdEncoding.EncodeToString(buf)
	return b64Pd
}


