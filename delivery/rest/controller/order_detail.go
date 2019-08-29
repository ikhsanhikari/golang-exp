package controller

import (
	"math"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/aging"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/device"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/installation"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order_detail"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/room"
)

func (c *Controller) insertOrderDetail(order order.Order, device device.Device, product product.Product, installation installation.Installation, room room.Room, aging aging.Aging, isAdmin bool) (err error) {
	var details = c.mappingDetailOrder(order.RoomQuantity, device, product, installation, room, aging)

	for _, detail := range details {
		insertDetail := order_detail.OrderDetail{
			OrderID:      order.OrderID,
			ItemType:     detail.ItemType,
			ItemID:       detail.ItemID,
			Description:  detail.Description,
			Amount:       detail.Amount,
			Quantity:     detail.Quantity,
			CreatedBy:    order.CreatedBy,
			LastUpdateBy: order.LastUpdateBy,
			ProjectID:    order.ProjectID,
		}

		err = c.orderDetail.Insert(&insertDetail, isAdmin)
		if err != nil {
			return err
		}
	}

	return
}

func (c *Controller) updateOrderDetail(order order.Order, device device.Device, product product.Product, installation installation.Installation, room room.Room, aging aging.Aging, isAdmin bool) (err error) {
	var details = c.mappingDetailOrder(order.RoomQuantity, device, product, installation, room, aging)

	for _, detail := range details {
		updateDetail := order_detail.OrderDetail{
			OrderID:      order.OrderID,
			ItemType:     detail.ItemType,
			ItemID:       detail.ItemID,
			Description:  detail.Description,
			Amount:       detail.Amount,
			Quantity:     detail.Quantity,
			CreatedBy:    order.CreatedBy,
			LastUpdateBy: order.LastUpdateBy,
			ProjectID:    order.ProjectID,
		}

		err = c.orderDetail.Update(&updateDetail, isAdmin)
		if err != nil {
			return err
		}
	}

	return
}

func (c *Controller) deleteOrderDetail(order order.Order, isAdmin bool) (err error) {

	deleteDetails := order_detail.OrderDetail{
		OrderID:      order.OrderID,
		ProjectID:    order.ProjectID,
		CreatedBy:    order.CreatedBy,
		LastUpdateBy: order.LastUpdateBy,
	}

	err = c.orderDetail.Delete(&deleteDetails, isAdmin)
	if err != nil {
		return err
	}

	return
}

func (c *Controller) mappingDetailOrder(roomQuantity int64, device device.Device, product product.Product, installation installation.Installation, room room.Room, aging aging.Aging) order_detail.Details {
	quantity := int64(1)

	details := order_detail.Details{
		{ItemType: "device", ItemID: device.ID, Description: device.Name, Amount: device.Price, Quantity: quantity},
		{ItemType: "product", ItemID: product.ProductID, Description: product.ProductName, Amount: product.Price, Quantity: quantity},
		{ItemType: "installation", ItemID: installation.ID, Description: installation.Name, Amount: installation.Price, Quantity: quantity},
		{ItemType: "aging", ItemID: aging.ID, Description: aging.Name, Amount: aging.Price, Quantity: quantity},
	}

	if room.ID != 0 {
		details = append(details, order_detail.Detail{
			ItemType: "room", ItemID: room.ID, Description: room.Name, Amount: (math.Ceil(float64(roomQuantity)*0.3) * room.Price), Quantity: roomQuantity,
		})
	}

	return details
}