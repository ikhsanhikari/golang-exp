package controller

import (
	"bytes"
	"database/sql"
	"net/http"
	"encoding/base64"
	"fmt"


	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/skip2/go-qrcode"
	"github.com/leekchan/accounting"
)

/*
return 0 to failed get Data
return 1 to failed get template
*/

func (c *Controller) handleBasePdf(id int64, userID string) string {
	var totPrice int64

	t, err := c.template.Get("pdf_invoice.tmpl")
	if err != nil {
		return "1"
	}

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
	gen.SetOutput(pdfBuffer)
	gen.AddPage(wkhtmltopdf.NewPageReader(buff))
	gen.Create()

	b := pdfBuffer.Bytes()
	b64Pdf := base64.StdEncoding.EncodeToString(b)

	return b64Pdf
}

func (c *Controller) handleBaseSertificatePdf(w http.ResponseWriter, r *http.Request){
	t, err := c.template.Get("pdf_sertificate.tmpl")
	if err != nil {
		return 
	}
	// var lid int64 = 26
	// venues, err := c.venue.SelectVenueByLisenceID(10, lid )
	// if err == sql.ErrNoRows {
	// 	c.reporter.Warningf("[handlePdf] Venue not found, err: %s", err.Error())
	// 	view.RenderJSONError(w, "Order not found", http.StatusNotFound)
	// 	return 
	// }
	// if err != nil && err != sql.ErrNoRows {
	// 	c.reporter.Errorf("[handlePdf] Failed get Venue, err: %s", err.Error())
	// 	view.RenderJSONError(w, "Failed get Venue", http.StatusNotFound)
	// 	return 
	// }
	sumorder, err := c.order.SelectSummaryOrderByID(153, 10, "RxHeyqVsEndVAUo2EBA4VBQWp207OO")//fmt.Sprintf("%v", userID))
    if err != nil {
        c.reporter.Errorf("[handleGetSumOrderByID] sum order not found, err: %s", err.Error())
        view.RenderJSONError(w, "Sum Order not found", http.StatusNotFound)
        return
    }

	templateData := map[string]interface{}{
		"VenueName"		:   sumorder.VenueName,
		"Address"		:   sumorder.VenueAddress,
		"Zip"			:   sumorder.VenueZip,
		"City"			:   sumorder.VenueCity,
		"Province"		:   sumorder.VenueProvince,
	}

	buff := bytes.NewBuffer([]byte{})
	err = t.Execute(buff, templateData)
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
	gen.Orientation.Set(wkhtmltopdf.OrientationLandscape)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"sertificate.pdf\"")
	gen.Create()

	// b := pdfBuffer.Bytes()
	// b64Pdf := base64.StdEncoding.EncodeToString(b)

	// return b64Pdf
}

func (c *Controller) handleGetPdf1(w http.ResponseWriter, r *http.Request) {
	view.RenderJSONData(w, c.handleBasePdf(153,"RxHeyqVsEndVAUo2EBA4VBQWp207OO"), http.StatusOK)
}

func (c *Controller) handleQrCode(licenseNumber string){

}



