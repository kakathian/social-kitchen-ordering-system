package kitchen

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	dispatchService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/dispatch"
	storageService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/storage"
	"time"

	"go.uber.org/zap"
)

var kitchenChannel chan model.Order = nil

func Start(noOfOrdersToRead int) {
	kitchenChannel = make(chan model.Order, noOfOrdersToRead)
	internalProcess()
}

func Process(orderReq model.Order) {
	// process order
	zap.S().Infof("Kitchen: Order '%s' (%s) getting processed", orderReq.Name, orderReq.ID)
	kitchenChannel <- orderReq
}

func internalProcess() {
	go func() {
		for {
			select {
			case orderReq, isOpen := <-kitchenChannel:
				if !isOpen {
					kitchenChannel = nil
					break
				}
				// Send order ready event
				zap.S().Infof("Kitchen: Order '%s'(%s) is ready at %s", orderReq.Name, orderReq.ID, time.Now())
				storageService.Process(orderReq)

				// Send order dispatch event
				zap.S().Infof("Kitchen: Order '%s'(%s) is ready for dispatch at %s", orderReq.Name, orderReq.ID, time.Now())
				dispatchService.Process(orderReq)
			}

			if kitchenChannel == nil {
				break
			}
		}
	}()
}

func Stop() {
	close(kitchenChannel)
}
