package payment

type TokenGenerator interface {
	GetAccessToken(pid int64) (string, error)
}

// Init is used to initialize room package
func Init(apiBaseURL string, tokenGenerator TokenGenerator) ICore {
	return &core{
		apiBaseURL:     apiBaseURL,
		tokenGenerator: tokenGenerator,
	}
}
