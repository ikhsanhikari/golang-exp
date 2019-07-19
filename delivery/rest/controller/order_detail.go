package controller

import (
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/aging"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/device"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/installation"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order_detail"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/room"
)

func (c *Controller) insertOrderDetail(order order.Order, device device.Device, product product.Product, installation installation.Installation, room room.Room, aging aging.Aging) (err error) {
	var details = []map[string]interface{}{
		{"itemType": "device", "itemID": device.ID, "description": device.Name, "amount": device.Price, "quantity": int64(1)},
		{"itemType": "product", "itemID": product.ProductID, "description": product.ProductName, "amount": product.Price, "quantity": int64(1)},
		{"itemType": "installation", "itemID": installation.ID, "description": installation.Description, "amount": installation.Price, "quantity": int64(1)},
		{"itemType": "aging", "itemID": aging.ID, "description": aging.Name, "amount": aging.Price, "quantity": int64(1)},
	}

	if room.ID != 0 {
		details = append(details, map[string]interface{}{
			"itemType": "room", "itemID": room.ID, "description": room.Name, "amount": room.Price, "quantity": order.RoomQuantity,
		})
	}

	for _, detail := range details {
		insertDetail := order_detail.OrderDetail{
			OrderID:      order.OrderID,
			ItemType:     detail["itemType"].(string),
			ItemID:       detail["itemID"].(int64),
			Description:  detail["description"].(string),
			Amount:       detail["amount"].(float64),
			Quantity:     detail["quantity"].(int64),
			CreatedBy:    order.CreatedBy,
			LastUpdateBy: order.LastUpdateBy,
			ProjectID:    order.ProjectID,
		}

		err = c.orderDetail.Insert(&insertDetail)
		if err != nil {
			return err
		}
	}

	return
}

func (c *Controller) updateOrderDetail(order order.Order, device device.Device, product product.Product, installation installation.Installation, room room.Room, aging aging.Aging) (err error) {
	var details = []map[string]interface{}{
		{"itemType": "device", "itemID": device.ID, "description": device.Name, "amount": device.Price, "quantity": int64(1)},
		{"itemType": "product", "itemID": product.ProductID, "description": product.ProductName, "amount": product.Price, "quantity": int64(1)},
		{"itemType": "installation", "itemID": installation.ID, "description": installation.Description, "amount": installation.Price, "quantity": int64(1)},
		{"itemType": "aging", "itemID": aging.ID, "description": aging.Name, "amount": aging.Price, "quantity": int64(1)},
	}

	if room.ID != 0 {
		details = append(details, map[string]interface{}{
			"itemType": "room", "itemID": room.ID, "description": room.Name, "amount": room.Price, "quantity": order.RoomQuantity,
		})
	}

	for _, detail := range details {
		updateDetail := order_detail.OrderDetail{
			OrderID:      order.OrderID,
			ItemType:     detail["itemType"].(string),
			ItemID:       detail["itemID"].(int64),
			Description:  detail["description"].(string),
			Amount:       detail["amount"].(float64),
			Quantity:     detail["quantity"].(int64),
			CreatedBy:    order.CreatedBy,
			LastUpdateBy: order.LastUpdateBy,
			ProjectID:    order.ProjectID,
		}

		err = c.orderDetail.Update(&updateDetail)
		if err != nil {
			return err
		}
	}

	return
}

func (c *Controller) deleteOrderDetail(order order.Order) (err error) {

	deleteDetails := order_detail.OrderDetail{
		OrderID:      order.OrderID,
		ProjectID:    order.ProjectID,
		CreatedBy:    order.CreatedBy,
		LastUpdateBy: order.LastUpdateBy,
	}

	err = c.orderDetail.Delete(&deleteDetails)
	if err != nil {
		return err
	}

	return
}
