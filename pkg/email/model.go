package email

type EmailRequest struct {
	Subject string `json:"subject"`
	To      string `json:"to"`
	HTML    string `json:"html"`
	From    string `json:"from"`
	Text    string `json:"text"`
}
