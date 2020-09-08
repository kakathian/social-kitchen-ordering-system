package sharedkitchenordersystem

import (
	"fmt"
	"math"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/order"
	dispatchService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/dispatch"
	kitchenService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/kitchen"
	storageService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/storage"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/supervisor"
	"time"

	"go.uber.org/zap"
)

// Initialize the application.
func Start(noOfOrdersToRead int) {
	// initliaze repos
	order.InitOrders()

	// start services
	supervisor.Start(noOfOrdersToRead)

	dispatchService.Start(noOfOrdersToRead)
	storageService.Start(noOfOrdersToRead)
	kitchenService.Start(noOfOrdersToRead)

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
				zap.S().Infof("Admin: Order '%s' (%s) received", order.Name, order.ID)
				orderReaderChannel <- order
			}

			time.Sleep(time.Second)
		}

		zap.S().Info("Admin: No more receiving Orders; kitchen closed")
		zap.S().Info("===============================================")

		// stop services
		//dispatchService.Stop()
		//storageService.Stop()
		//kitchenService.Stop()

		close(orderReaderChannel)
	}(order.OrdersData)

	for {
		select {
		case orderReq, isOpen := <-orderReaderChannel:
			if !isOpen {
				orderReaderChannel = nil
				break
			}

			zap.S().Infof("Admin: Order '%s'(%s) is being sent to kitchen at %s", orderReq.Name, orderReq.ID, time.Now())
			supervisor.KitchenChannel <- orderReq
		}

		if orderReaderChannel == nil {
			break
		}
	}

	time.Sleep(10 * time.Second)
	supervisor.Report.GenerateReport()
}
