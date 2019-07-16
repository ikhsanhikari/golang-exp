package controller

import (
	"bytes"
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

func (c *Controller) handlePdf(w http.ResponseWriter, r *http.Request) {
	t, err := c.template.Get("pdf_invoice.tmpl")
	if err != nil {
		view.RenderJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buff := bytes.NewBuffer([]byte{})
	err = t.Execute(buff, map[string]string{"Name": "Risal"})
	if err != nil {
		view.RenderJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gen, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		view.RenderJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	gen.SetOutput(w)
	gen.AddPage(wkhtmltopdf.NewPageReader(buff))

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"invoice.pdf\"")
	gen.Create()
}
