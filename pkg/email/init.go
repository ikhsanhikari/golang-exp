package email

type TokenGeneratorEmail interface {
	GetAccessToken(pid int64) (string, error)
}

// Init is used to initialize room package
func Init(apiBaseURL string, urlQrCode string, tokenGeneratorEmail TokenGeneratorEmail) ICore {
	return &core{
		urlQrCode:     urlQrCode,
		apiBaseURL:     apiBaseURL,
		tokenGeneratorEmail: tokenGeneratorEmail,
	}
}
