package email

type EmailRequest struct {
	Subject    string      `json:"subject"`
	To         string      `json:"to"`
	HTML       string      `json:"html"`
	From       string      `json:"from"`
	Text       string      `json:"text"`
	Attachment Attachments `json:"attachments"`
}
type Attachment struct {
	Content     string `json:"content"`
	Filename    string `json:"filename"`
	Type        string `json:"type"`
	Disposition string `json:"disposition"`
	ContentID   string `json:"contentId"`
}

type Attachments []Attachment
