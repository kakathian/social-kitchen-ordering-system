package repository

import (
	"fmt"
	"sharedkitchenordersystem/model"
	"sharedkitchenordersystem/utility"
)

var ordersData []model.Order

func InitOrders() {
	ordersData = []model.Order{}
	utility.ReadFile("orders.json", &ordersData)
	fmt.Println(len(ordersData))
}
