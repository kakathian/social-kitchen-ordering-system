package service

import (
	"fmt"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	orderRepo "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/order"
)

func ProcessOrders() model.Order {
	orders []model.Order  = orderRepo.OrdersData
	for _, order := range orders {
		fmt.Println(order.Id)
	}	
}