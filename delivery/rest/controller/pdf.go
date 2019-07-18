package controller

import (
	"bytes"
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"github.com/leekchan/accounting"

	//"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

func (c *Controller) handlePdf(w http.ResponseWriter, r *http.Request) {
	t, err := c.template.Get("pdf_invoice.tmpl")
	if err != nil {
		view.RenderJSONError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// order, err := c.order.Get(111, 10, "uid")
	// if err == sql.ErrNoRows {
	// 	c.reporter.Warningf("[handlePatchOrder] order not found, err: %s", err.Error())
	// 	view.RenderJSONError(w, "Order not found", http.StatusNotFound)
	// 	return
	// }
	// if err != nil && err != sql.ErrNoRows {
	// 	c.reporter.Errorf("[handlePatchOrder] Failed get order, err: %s", err.Error())
	// 	view.RenderJSONError(w, "Failed get order", http.StatusInternalServerError)
	// 	return
	// }

	templateData := map[string]interface{}{
		"CreatedAt":         "2020-01-01",
		"OrderNumber":       "XX-12",
		"CustomerReference": "XX-123-123132",
		"BuyerName":         "PT Liga Inggris",
		"BuyerAddress":      "Jalan EPL, No.153, Surabaya 17865489",
		"Items": []map[string]interface{}{
			{
				"Quantity":           2,
				"ProductName":        "Product One",
				"ProductDescription": "ProductDescription",
				"ProductPrice":       "100.000",
				"TotalPrice":         "200.000",
			},
			{
				"Quantity":           1,
				"ProductName":        "Product Two",
				"ProductDescription": "Product Description Two",
				"ProductPrice":       "100.000",
				"TotalPrice":         "100.000",
			},
		},
		"Subtotal":   "300.000",
		"Total":      "300.000",
		"BalanceDue": "300.000",
	}

	_ = accounting.DefaultAccounting

	// now := order.CreatedAt
	// ac := accounting.Accounting{Precision: 2, Thousand: ".", Decimal: ","}
	// totPrice := ac.FormatMoney(order.TotalPrice)
	// //if len(data) > 0 {
	// // var oDatas orderDatass
	// // for _, order := range data {
	// var orderDatas orderData
	// productParam, err := c.product.Get(10, order.ProductID)
	// fmt.Println(productParam)
	// uPrice := ac.FormatMoney(productParam.Price)
	// orderDatas.OrderNumber = order.OrderNumber
	// orderDatas.BuyerID = order.OrderNumber
	// orderDatas.VenueID = order.VenueID
	// orderDatas.DeviceID = order.DeviceID
	// orderDatas.ProductID = order.ProductID
	// orderDatas.InstallationID = order.InstallationID
	// orderDatas.Quantity = order.Quantity
	// orderDatas.AgingID = order.AgingID
	// orderDatas.RoomID = order.RoomID
	// orderDatas.RoomQuantity = order.RoomQuantity
	// orderDatas.TotalPrice = totPrice
	// orderDatas.PaymentMethodID = order.PaymentMethodID
	// orderDatas.PaymentFee = order.PaymentFee
	// orderDatas.Status = order.Status
	// orderDatas.CreatedAt = now.Format("2006-01-02")
	// orderDatas.LastUpdateBy = order.LastUpdateBy
	// orderDatas.DeletedAt = order.DeletedAt
	// orderDatas.Email = order.Email
	// orderDatas.ProductName = productParam.ProductName
	// orderDatas.ProductPrice = uPrice
	// orderDatas.Description = productParam.Description
	// orderDatas.DeviceName = ""       //order.DeviceName
	// orderDatas.InstallationName = "" //order.InstallationName
	// orderDatas.AgingName = ""        //order.AgingName
	// orderDatas.RoomName = ""         //order.RoomName
	// // 	oDatas = append(oDatas,orderDatas)
	// // }
	// //}

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

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "filename=\"invoice.pdf\"")
	gen.Create()
}
