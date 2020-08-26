package repository

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sharedkitchenordersystem/model"
	"sharedkitchenordersystem/utility"
)

var OrdersData []model.Order

func InitOrders() {
	OrdersData = []model.Order{}
	_, b, _, _ := runtime.Caller(0)

	utility.ReadFile(filepath.Dir(b)+"/orders.json", OrdersData)
	fmt.Println(len(OrdersData))
}
