package service

import (
	"fmt"
	orderRepo "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/order"
)

// ProcessOrders .
func ProcessOrders() {
	for _, order := range orderRepo.OrdersData {
		fmt.Println(order.ID)
	}
}
