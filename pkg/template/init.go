package template

// Init is used to initialize room package
func Init(urlQrCode string) ICore {
	return &core{
		urlQrCode:     urlQrCode,
	}
}
