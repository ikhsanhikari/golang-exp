package controller

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
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
		gen.MarginBottom.Set(0)
		gen.MarginTop.Set(0)
		gen.MarginLeft.Set(0)
		gen.MarginRight.Set(0)
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

	sumorder, err := c.order.SelectSummaryOrderByID(orderid, 10, userID)
	if err != nil {
		c.reporter.Errorf("[handleSertificatePDF] sum order not found, err: %s", err.Error())
		return "0", sumorder
	}
	if sumorder.LicenseNumber == "" {
		c.reporter.Errorf("[handleSertificatePDF] License number not found, err: %s", err.Error())
		return "0", sumorder
	}

	b64Png, backBase64 := c.email.GetBase64Png(sumorder.LicenseNumber)
	if b64Png == "0" && backBase64 == "0" {
		c.reporter.Errorf("[handleSertificatePDF] Error base64 from image")
		return "0", sumorder
	}

	templateData := map[string]interface{}{
		"VenueName":  sumorder.VenueName,
		"Address":    sumorder.VenueAddress,
		"Zip":        sumorder.VenueZip,
		"City":       sumorder.VenueCity,
		"Province":   sumorder.VenueProvince,
		"QrBase64":   b64Png,
		"Background": backBase64,
	}
	b64SertificatePdf := c.handleBasePdf(templateData, t, pdf, "Landscape")
	return b64SertificatePdf, sumorder

}

func (c *Controller) handleGetHtmlBodyCert(venueName string) string {

	file, err := os.Open("file/img_email_cert/artboard-background.png")
	if err != nil {
		return "0"
	}
	defer file.Close()
	fInfo, _ := file.Stat()
	var size int64 = fInfo.Size()
	buf := make([]byte, size)
	fReader := bufio.NewReader(file)
	fReader.Read(buf)
	artboardBackground := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-combined-shape-62@2x.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardCombinedShape62 := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-ellipse-1@2x.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardEllipse1 := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-layer-1.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardLayer1 := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-layer-2@2x.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardLayer2 := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-layer-3@2x.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardLayer3 := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-rectangle-3-1.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardRectangle31 := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-rectangle-3-2.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardRectangle32 := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-rectangle-3.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardRectangle3 := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-shape-1-copy.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardShape1Copy := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-shape-1.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardShape1 := base64.StdEncoding.EncodeToString(buf)

	file, err = os.Open("file/img_email_cert/artboard-shape-2.png")
	if err != nil {
		return "0"
	}
	fInfo, _ = file.Stat()
	size = fInfo.Size()
	buf = make([]byte, size)
	fReader = bufio.NewReader(file)
	fReader.Read(buf)
	artboardShape2 := base64.StdEncoding.EncodeToString(buf)

	templateData := map[string]interface{}{
		"Artboardbackground":        artboardBackground,
		"Artboardshape2":            artboardShape2,
		"Artboardshape1":            artboardShape1,
		"Artboardlayer1":            artboardLayer1,
		"Artboardshape1copy":        artboardShape1Copy,
		"VenueName":                 venueName,
		"Artboardrectangle3":        artboardRectangle3,
		"Artboardellipse12x":        artboardEllipse1,
		"Artboardlayer22x":          artboardLayer2,
		"Artboardrectangle31":       artboardRectangle31,
		"Artboardlayer32x":          artboardLayer3,
		"Artboardrectangle32":       artboardRectangle32,
		"Artboardcombinedshape622x": artboardCombinedShape62,
	}

	t, err := c.template.Get("email_sertificate.tmpl")
	if err != nil {
		return "1"
	}

	buff := bytes.NewBuffer([]byte{})
	err = t.Execute(buff, templateData)
	if err != nil {
		c.reporter.Errorf("[handleGetHtmlBodyCert] Failed execute html, err: %s", err.Error())
		return "1"
	}

	return buff.String()

}

func (c *Controller) handleGetPdf1(w http.ResponseWriter, r *http.Request) {
	a, _ := c.handleGetDataSertificate(150, "kDQ2IAaHPZ8MTkqNS24zJPKu9MSLBo")
	view.RenderJSONData(w, a, http.StatusOK)
}
