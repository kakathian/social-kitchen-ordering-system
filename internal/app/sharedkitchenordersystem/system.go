package sharedkitchenordersystem

import (
	"fmt"
	"go.uber.org/zap"
	"math"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/order"
	"time"
)

// Initialize the application.
func Initialize(noOfOrdersToRead int) {
	// initliaze repos
	order.InitOrders()

	orderReaderChannel := make(chan model.Order, noOfOrdersToRead)

	go func(ordersData []model.Order) {

		fmt.Println("entered")
		if len(ordersData) == 0 {
			return
		}

		end := 0
		for i := 0; i < len(ordersData); i += noOfOrdersToRead {
			end = int(math.Min(float64(i+noOfOrdersToRead), float64(len(ordersData))))
			for _, order := range ordersData[i:end] {
				zap.S().Infof("Order '%s' (%s) received", order.Name, order.ID)
				orderReaderChannel <- order
			}

			time.Sleep(time.Second)
		}

		close(orderReaderChannel)
		zap.S().Info("No more receiving Orders; kitchen closed")
	}(order.OrdersData)

	for {
		select {
		case orderReq, isOpen := <-orderReaderChannel:
			if !isOpen {
				orderReaderChannel = nil
				break
			}
			zap.S().Infof("Order '%s'(%s) started processing at %s", orderReq.Name, orderReq.ID, time.Now())
		}

		if orderReaderChannel == nil {
			break
		}
	}
}
