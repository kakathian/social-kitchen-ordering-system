package kitchen

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	dispatchService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/dispatch"
	storageService "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/storage"
	"sharedkitchenordersystem/internal/pkg"
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
				shelfItem := model.ShelfItem{
					Order:        orderReq,
					CreatedTime:  time.Now(),
					MaxLifeTimeS: pkg.CalculateMaxAge(orderReq.ShelfLife, orderReq.DecayRate, 1), // assume 1 for now, then change dynamically later
				}
				zap.S().Infof("Kitchen: Order '%s'(%s) is ready and expires in %d(s)", orderReq.Name, orderReq.ID, shelfItem.MaxLifeTimeS)
				storageService.Process(shelfItem)

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
