package dispatch

import (
	"math/rand"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	repo "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/shelf"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/supervisor"
	"strings"
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
			case orderReq, isOpen := <-supervisor.DispatchChannel:
				if !isOpen {
					supervisor.DispatchChannel = nil
					break
				}
				// Send order ready event
				rand.Seed(time.Now().UnixNano())

				// Courier arrived randomly after this time
				time.Sleep(time.Duration(rand.Intn(6-2+1)+2) * time.Second)
				//time.Sleep(6 * time.Second)

				// Courier picking up the order
				shelf, err := repo.ShelfFactory(orderReq.Temp)

				if err != nil {
					zap.S().Infof("Dispatch: Invalid Order '%s'(%s); ignored unknown order item temperature '%s'", orderReq.ID, orderReq.Name, orderReq.Temp)
					continue
				}

				// If item not available in normal racks, check in overflow rack
				isPresent := shelf.IsPresent(orderReq.ID)
				isOrderDispatched := false
				pickedUpShelfType := ""

				// Check if present in normal shelves
				if isPresent {
					shelf.Delete(orderReq.ID)
					isOrderDispatched = true
					pickedUpShelfType = orderReq.Temp
					zap.S().Infof("Dispatch: Order '%s'(%s) removed from shelf '%s' by courier", orderReq.ID, orderReq.Name, orderReq.Temp)
				} else {
					// Check if present in overflow shelf
					overflownShelf := repo.OverflowShelf[strings.ToLower(orderReq.Temp)]
					if isPresent = overflownShelf.IsPresent(orderReq.ID); isPresent {
						overflownShelf.Delete(orderReq.ID)
						isOrderDispatched = true
						pickedUpShelfType = model.OVERFLOW
					}
				}

				if isOrderDispatched {
					zap.S().Infof("Dispatch: Courier picked up Order '%s'(%s) from '%s' shelf ", orderReq.Name, orderReq.ID, pickedUpShelfType)

					// Send OrderStatus event
					supervisor.SupervisorChannel <- model.OrderStatus{OrderId: orderReq.ID, Status: model.ORDER_PICKED}

					// Once courier picked up the order (shelf item), send new space available event
					if pickedUpShelfType != model.OVERFLOW {
						supervisor.NewSpaceAvailableChannel <- orderReq.Temp
						zap.S().Infof("Dispatch: New space available in shelf for '%s'", orderReq.Temp)
					}
				} else {
					// Order could not be found, probably discarded - should be confirmed discarded/expired with supervisor
					var status string = "Not Available"
					if supervisor.Report.IsTrashed(orderReq.ID) {
						status = model.ORDER_EXPIRED
					} else if supervisor.Report.IsEvicted(orderReq.ID) {
						status = model.ORDER_EVICTED
					}

					zap.S().Infof("Dispatch: Courier could not find the Order '%s'(%s) in shelves; it is '%s'", orderReq.Name, orderReq.ID, status)
				}
			}

			if supervisor.DispatchChannel == nil {
				break
			}
		}
	}()
}
