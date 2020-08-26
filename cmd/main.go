package main

import (
	"fmt"
	"sharedkitchenordersystem/model"
	"sharedkitchenordersystem/repository"
)

var ShelvesCapacity model.ShelfCapacity

func main() {
	fmt.Println("hello")
	fmt.Println(model.ShelvesCapacity.Hot)
	repository.InitOrders()

	for i := 0; i < len(repository.OrdersData); i++ {
		fmt.Println("Order Id: ", repository.OrdersData[i].Id)
		fmt.Println("Order Name: ", repository.OrdersData[i].Name)
	}
}
