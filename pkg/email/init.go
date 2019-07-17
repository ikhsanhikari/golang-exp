package email

type TokenGeneratorEmail interface {
	GetAccessToken(pid int64) (string, error)
}

// Init is used to initialize room package
func Init(apiBaseURL string, tokenGeneratorEmail TokenGeneratorEmail) ICore {
	return &core{
		apiBaseURL:     apiBaseURL,
		tokenGeneratorEmail: tokenGeneratorEmail,
	}
}
