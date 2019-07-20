package controller

type emailRequest struct {
	Subject     string      `json:"subject"`
	To          string      `json:"to"`
	HTML        string      `json:"html"`
	From        string      `json:"from"`
	Text        string      `json:"text"`
	Attachments Attachments `json:"attachments"`
}

type AttachmentReq struct {
	Content     string `json:"content"`
	Filename    string `json:"filename"`
	Type        string `json:"type"`
	Disposition string `json:"disposition"`
	ContentID   string `json:"contentId"`
}

type Attachments []AttachmentReq
