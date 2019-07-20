package template

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	qrcode "github.com/skip2/go-qrcode"
)

type ICore interface {
	Get(name string) (*template.Template, error)
	GetBase64Png(licenseNum string) string
}

type core struct {
	t         map[string]*template.Template
	mux       sync.RWMutex
	urlQrCode string
}

// ErrTemplateNotFound ...
var ErrTemplateNotFound = errors.New("Template is not found")

// New ...
func New(paths ...string) ICore {
	var ts = core{t: make(map[string]*template.Template, 0)}
	for _, path := range paths {
		f, err := os.Stat(path)
		if err != nil {
			panic(err)
		}

		if !f.IsDir() {
			panic(errors.New("You must pass directory path"))
		}

		err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if !isTemplate(path) {
				return nil
			}

			t, err := template.ParseFiles(path)
			if err != nil {
				return err
			}

			ts.t[t.Name()] = t
			return nil
		})

		if err != nil {
			panic(err)
		}
	}

	return &ts
}

// Get ...
func (c *core) Get(name string) (*template.Template, error) {
	c.mux.RLock()
	t, ok := c.t[name]
	c.mux.RUnlock()
	if ok {
		return t.Clone()
	}
	return nil, ErrTemplateNotFound
}

func isTemplate(path string) bool {
	return filepath.Ext(path) == ".tmpl"
}

func (c *core) GetBase64Png(licenseNum string) string {
	fmt.Println(c.urlQrCode)
	fmt.Println(licenseNum)
	qrCode := c.urlQrCode + licenseNum
	fmt.Println(qrCode)
	var png []byte
	png, _ = qrcode.Encode(qrCode, qrcode.Medium, 256)
	fmt.Println(png)
	b64Png := base64.StdEncoding.EncodeToString(png)
	fmt.Println(b64Png)

	return b64Png
}
