package order

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	util "sharedkitchenordersystem/pkg"
)

var OrdersData []model.Order

// InitOrders initializes and reads orders data from orders json file
func InitOrders() {
	OrdersData = []model.Order{}
	_, b, _, _ := runtime.Caller(0)
	err := util.ReadFile(filepath.Dir(b)+"/orders.json", &OrdersData)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Orders data read: length is %d", len(OrdersData))
}
