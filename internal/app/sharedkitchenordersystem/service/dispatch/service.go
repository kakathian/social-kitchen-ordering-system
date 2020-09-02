package dispatch

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"time"

	"go.uber.org/zap"
)

var dispatchChannel chan model.Order = nil

func Start(noOfOrdersToRead int) {
	dispatchChannel = make(chan model.Order, noOfOrdersToRead)
	internalProcess()
}

func Process(orderReq model.Order) {
	// process order
	zap.S().Infof("Dispatch: Order '%s' (%s) getting dispatched", orderReq.Name, orderReq.ID)
	dispatchChannel <- orderReq
}

func internalProcess() {
	go func() {
		for {
			select {
			case orderReq, isOpen := <-dispatchChannel:
				if !isOpen {
					dispatchChannel = nil
					break
				}
				// Send order ready event
				zap.S().Infof("Dispatch: Order '%s'(%s) notified to get dispatched at %s", orderReq.Name, orderReq.ID, time.Now())
			}

			if dispatchChannel == nil {
				break
			}
		}
	}()
}

func Stop() {
	close(dispatchChannel)
}
