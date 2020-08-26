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
}
