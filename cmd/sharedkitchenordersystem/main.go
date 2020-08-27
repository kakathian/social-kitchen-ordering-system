package main

import (
	"fmt"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/order"

	"go.uber.org/zap"
)

var ShelvesCapacity model.ShelfCapacity

func main() {
	// Create a zap logger with appropriate configuration.
	logger := createZapLogger()
	// Replace the global logger with created logger. After this
	// any package in the process can use zap.L() or zap.S()
	zap.ReplaceGlobals(logger)

	zap.S().Info("Starting main method")

	// initliaze repos
	order.InitOrders()

	for i := 0; i < len(order.OrdersData); i++ {
		fmt.Println("Order Id: ", repository.OrdersData[i].Id)
		fmt.Println("Order Name: ", repository.OrdersData[i].Name)
	}
}

func createZapLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.S().Fatal(err)
	}
	return logger
}
