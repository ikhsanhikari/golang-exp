package controller

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/leekchan/accounting"
)

/*
return 0 to failed get Data
return 1 to failed get template
*/

func (c *Controller) handleBasePdf(templateData map[string]interface{}, tmp string, nameFile string, orientation string) string {
	t, err := c.template.Get(tmp)
	if err != nil {
		return "1"
	}

	buff := bytes.NewBuffer([]byte{})
	err = t.Execute(buff, templateData)
	if err != nil {
		c.reporter.Errorf("[handlePDF] Failed execute pdf, err: %s", err.Error())
		return "1"
	}

	pdfBuffer := bytes.NewBuffer([]byte{})
	gen, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		c.reporter.Errorf("[handlePDF] Failed generate pdf, err: %s", err.Error())
		return "1"
	}

	if orientation == "Landscape" {
		gen.Orientation.Set(wkhtmltopdf.OrientationLandscape)
	}
	gen.SetOutput(pdfBuffer)
	gen.AddPage(wkhtmltopdf.NewPageReader(buff))
	gen.Create()

	b := pdfBuffer.Bytes()
	b64Pdf := base64.StdEncoding.EncodeToString(b)

	return b64Pdf
}

func (c *Controller) handleGetDataInvoice(id int64, userID string) string {
	var totPrice int64

	t := "pdf_invoice.tmpl"
	pdf := "invoice.pdf"

	order, err := c.order.Get(id, 10, fmt.Sprintf("%v", userID))
	if err == sql.ErrNoRows {
		c.reporter.Warningf("[handlePatchPDF] order not found, err: %s", err.Error())
		return "0"
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchPDF] Failed get order, err: %s", err.Error())
		return "0"
	}

	orderDetail, err := c.orderDetail.Get(id, 10, fmt.Sprintf("%v", userID))
	if err == sql.ErrNoRows {
		c.reporter.Warningf("[handlePatchPDF] orderDetail not found, err: %s", err.Error())
		return "0"
	}

	ac := accounting.Accounting{Precision: 2, Thousand: ".", Decimal: ","}

	items := make([]map[string]interface{}, 0, len(orderDetail))
	for _, v := range orderDetail {
		typePrice := v.Quantity * int64(v.Amount)
		items = append(items, map[string]interface{}{
			"Quantity":     v.Quantity,
			"ProductName":  v.Description,
			"ProductPrice": ac.FormatMoney(v.Amount),
			"TotalPrice":   ac.FormatMoney(typePrice),
		})
		totPrice = totPrice + typePrice
	}

	templateData := map[string]interface{}{
		"CreatedAt":         order.CreatedAt.Format("2006-01-02"),
		"OrderNumber":       order.OrderID,
		"CustomerReference": "",
		"BuyerName":         "PT Liga Inggris",
		"BuyerAddress":      "Jalan EPL, No.153, Surabaya 17865489",
		"Items":             items,
		"Subtotal":          ac.FormatMoney(totPrice),
		"Total":             ac.FormatMoney(totPrice),
		"BalanceDue":        ac.FormatMoney(totPrice),
	}

	b64InvoicePdf := c.handleBasePdf(templateData, t, pdf, "Potrait")

	return b64InvoicePdf
}

func (c *Controller) handleGetDataSertificate(orderid int64, userID string) (string, order.SummaryOrder) {

	t := "pdf_sertificate.tmpl"
	pdf := "sertificate.pdf"

	sumorder, err := c.order.SelectSummaryOrderByID(orderid, 10, userID) //fmt.Sprintf("%v", userID))
	if err != nil {
		c.reporter.Errorf("[handleSertificatePDF] sum order not found, err: %s", err.Error())
		return "0", sumorder
	}
	if sumorder.LicenseNumber == "" {
		c.reporter.Errorf("[handleSertificatePDF] License number not found, err: %s", err.Error())
		return "0", sumorder
	}

	b64Png := c.email.GetBase64Png(sumorder.LicenseNumber)

	templateData := map[string]interface{}{
		"VenueName": sumorder.VenueName,
		"Address":   sumorder.VenueAddress,
		"Zip":       sumorder.VenueZip,
		"City":      sumorder.VenueCity,
		"Province":  sumorder.VenueProvince,
		"QrBase64":  b64Png,
	}
	b64SertificatePdf := c.handleBasePdf(templateData, t, pdf, "Landscape")
	return b64SertificatePdf, sumorder

}
