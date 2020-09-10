package kitchen

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/supervisor"
	"sharedkitchenordersystem/internal/pkg"
	"time"

	"go.uber.org/zap"
)

func Start(noOfOrdersToRead int) {
	internalProcess()
}

func internalProcess() {
	go func() {
		for {
			select {
			case orderReqs, more := <-supervisor.KitchenChannel:
				if !more {
					supervisor.KitchenChannel = nil
					break
				}
				go func(orderReqs []model.Order) {
					for _, orderReq := range orderReqs {
						zap.S().Infof("Kitchen: Order '%s' (%s) getting processed", orderReq.Name, orderReq.ID)

						// Send order status event
						supervisor.SupervisorChannel <- model.OrderStatus{OrderId: orderReq.ID, Status: model.ORDER_RECEIVED}

						// Send order ready event
						shelfItem := model.ShelfItem{
							Order:        orderReq,
							CreatedTime:  time.Now(),
							MaxLifeTimeS: pkg.CalculateMaxAge(orderReq.ShelfLife, orderReq.DecayRate, 1), // assume 1 for now, then change dynamically later
						}
						zap.S().Infof("Kitchen: Order '%s'(%s) is ready and expires in %d(s)", orderReq.Name, orderReq.ID, shelfItem.MaxLifeTimeS)

						// Send OrderStatus event
						supervisor.SupervisorChannel <- model.OrderStatus{OrderId: orderReq.ID, Status: model.ORDER_PROCESSED}

						// Send StoreOrder event
						zap.S().Infof("Kitchen: Order '%s' (%s) sent to Storage to get stored", shelfItem.Order.Name, shelfItem.Order.ID)
						supervisor.StorageChannel <- shelfItem

						// Send InitiateDispatcher event
						zap.S().Infof("Kitchen: Order '%s'(%s) is ready for dispatch and sent to Dispatch at %s", orderReq.Name, orderReq.ID, time.Now())
						supervisor.DispatchChannel <- orderReq
					}
				}(orderReqs)
			}

			if supervisor.KitchenChannel == nil {
				break
			}
		}
	}()
}
