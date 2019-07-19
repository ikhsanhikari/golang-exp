package controller

import (
	"fmt"
	"bytes"
	"database/sql"
	"encoding/base64"
	
	"github.com/leekchan/accounting"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

/* 
return 0 to failed get Data
return 1 to failed get template
*/

func (c *Controller) handleBasePdf(id int64, userID string) string  {
	var totPrice int64

	t, err := c.template.Get("pdf_invoice.tmpl")
	if err != nil {
		return "1"
	}

	order, err := c.order.Get(id, 10, fmt.Sprintf("%v",userID))
	if err == sql.ErrNoRows {
		c.reporter.Warningf("[handlePatchOrder] order not found, err: %s", err.Error())
		return "0"
	}
	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchOrder] Failed get order, err: %s", err.Error())
		return "0"
	}
	fmt.Println(order)

	orderDetail, err := c.orderDetail.Get(id, 10, fmt.Sprintf("%v",userID))
	if err == sql.ErrNoRows {
		c.reporter.Warningf("[handlePatchOrderDetail] orderDetail not found, err: %s", err.Error())
		return "0"
	}

	ac := accounting.Accounting{ Precision: 2, Thousand: ".", Decimal: ","}

	items := make([]map[string]interface{}, 0, len(orderDetail))
	for _, v := range orderDetail {
		typePrice := v.Quantity * int64(v.Amount)
		items = append(items, map[string]interface{}{
			"Quantity"			: v.Quantity,
			"ProductName"		: v.Description,
			"ProductPrice"		: ac.FormatMoney(v.Amount),
			"TotalPrice"		: ac.FormatMoney(typePrice),		
		})
		totPrice = totPrice + typePrice
	}

	templateData := map[string]interface{}{
		"CreatedAt"			: order.CreatedAt.Format("2006-01-02"),
		"OrderNumber"		: order.OrderID,
		"CustomerReference" : "",
		"BuyerName"			: "PT Liga Inggris",
		"BuyerAddress"		: "Jalan EPL, No.153, Surabaya 17865489",
		"Items"				: items,
		"Subtotal"			: ac.FormatMoney(totPrice),
		"Total"				: ac.FormatMoney(totPrice),
		"BalanceDue"		: ac.FormatMoney(totPrice),
	}

	buff := bytes.NewBuffer([]byte{})
	err = t.Execute(buff, templateData)
	if err != nil {
		return "1"
	}

	pdfBuffer := bytes.NewBuffer([]byte{})
    gen, err := wkhtmltopdf.NewPDFGenerator()
    if err != nil {
        return "1"
    }
    gen.SetOutput(pdfBuffer)
    gen.AddPage(wkhtmltopdf.NewPageReader(buff))
    gen.Create()

    b := pdfBuffer.Bytes()
	b64Pdf := base64.StdEncoding.EncodeToString(b)
	fmt.Println(b64Pdf)
	
	return b64Pdf
}

