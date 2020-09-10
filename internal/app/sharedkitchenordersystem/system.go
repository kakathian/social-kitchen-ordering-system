package sharedkitchenordersystem

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/order"
	dispatchService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/dispatch"
	kitchenService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/kitchen"
	storageService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/storage"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/supervisor"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Initialize the application.
func Start(noOfOrdersToRead int) {

	listenToSystemCloseSignal()

	// initliaze repos
	order.InitOrders()

	// start services
	supervisor.Start(noOfOrdersToRead)

	dispatchService.Start(noOfOrdersToRead)
	storageService.Start(noOfOrdersToRead)
	kitchenService.Start(noOfOrdersToRead)

	orderReaderChannel := make(chan []model.Order, noOfOrdersToRead)

	go func(ordersData []model.Order) {

		fmt.Println("entered")
		if len(ordersData) == 0 {
			return
		}

		end := 0
		for i := 0; i < len(ordersData); i += noOfOrdersToRead {
			end = int(math.Min(float64(i+noOfOrdersToRead), float64(len(ordersData))))
			/*for _, order := range ordersData[i:end] {
				zap.S().Infof("Admin: Order '%s' (%s) received", order.Name, order.ID)
				orderReaderChannel <- order
			}*/

			orderReaderChannel <- ordersData[i:end]

			time.Sleep(time.Second)
		}

		zap.S().Info("Admin: No more receiving Orders; kitchen closed")
		zap.S().Info("===============================================")

		// stop services
		//dispatchService.Stop()
		//storageService.Stop()
		//kitchenService.Stop()
		// close(orderReaderChannel)
	}(order.OrdersData)

	for {
		select {
		case orderReqs, isOpen := <-orderReaderChannel:
			if !isOpen {
				orderReaderChannel = nil
				break
			}

			zap.S().Infof("Admin: Received number of orders '%d' and are being sent to kitchen at %s", len(orderReqs), time.Now())
			supervisor.KitchenChannel <- orderReqs
		}

		if orderReaderChannel == nil {
			break
		}
	}
}

// ListenToSystemCloseSignal listens to OS interrupt signal. Cleans resources and prints orders status report
func listenToSystemCloseSignal() {
	appCloseListener := make(chan os.Signal)
	signal.Notify(appCloseListener, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-appCloseListener
		zap.S().Infof("Admin: Received termination signal (%s).Printing order status report before closing....", "Ctrl+C")
		supervisor.Report.GenerateReport()
		zap.S().Info("Admin: Cleaning resources....")
		supervisor.CloseAll()
		zap.S().Info("----------------------Application shutting down----------------------")

		os.Exit(0)
	}()
}
