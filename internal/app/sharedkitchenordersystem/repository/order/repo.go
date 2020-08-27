package order

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	util "sharedkitchenordersystem/pkg"
)

var OrdersData []model.Order

func InitOrders() {
	OrdersData = []model.Order{}
	_, b, _, _ := runtime.Caller(0)

	err := util.ReadFile(filepath.Dir(b)+"/orders.json", &OrdersData)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(len(OrdersData))
}
