package payment

type Payment struct {
	ResponseType    string            `json:"responseType"`
	HTMLRedirection string            `json:"htmlRedirection"`
	PaymentData     PaymentAttributes `json:"paymentData"`
}

type PaymentAttributes struct {
	URL string `json:"url"`
}
