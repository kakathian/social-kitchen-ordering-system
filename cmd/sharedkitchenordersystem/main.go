package main

import (
	"fmt"
	"go.uber.org/zap"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/order"
)

func main() {
	// Create a zap logger with appropriate configuration.
	logger := createZapLogger()
	// Replace the global logger with created logger. After this
	// any package in the process can use zap.L() or zap.S()
	zap.ReplaceGlobals(logger)

	zap.S().Info("Starting main method")

	// initliaze repos
	order.InitOrders()

	// noOfOrdersToRead := 2

	// read orders
	// orderReader := make(chan model.OrderRequest, noOfOrdersToRead)

	go func(ordersData []model.Order) {

		fmt.Println("entered")
		// if len(ordersData) == 0 {
		// 	return
		// }
		// end := 0
		// for i := 0; i < len(ordersData); i += noOfOrdersToRead {

		// 	if end > len(ordersData) {
		// 		end = len(ordersData)
		// 	}

		// 	end += noOfOrdersToRead

		// 	for _, order := range ordersData[i:end] {

		// 		zap.S().Info("Order sent")
		// 		orderReader <- model.OrderRequest{
		// 			Order: order,
		// 			Time:  time.Now(),
		// 		}
		// 	}

		// 	time.Sleep(time.Second)
		// }
	}(order.OrdersData)

	go func() {
		fmt.Println("left")
		// for {
		// 	select {
		// 	case orderReq := <-orderReader:
		// 		zap.S().Info("Order received " + orderReq.Order.Id)
		// 	}
		// }
	}()

	/*for i := 0; i < len(order.OrdersData); i++ {
		fmt.Println("Order Id: ", repository.OrdersData[i].Id)
		fmt.Println("Order Name: ", repository.OrdersData[i].Name)
	}*/
}

func createZapLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.S().Fatal(err)
	}
	return logger
}
