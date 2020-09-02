package storage

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"time"

	"go.uber.org/zap"
)

// ProcessOrders .
var storageChannel chan model.Order = nil

func Start(noOfOrdersToRead int) {
	storageChannel = make(chan model.Order, noOfOrdersToRead)
	internalProcess()
}

func Process(orderReq model.Order) {
	// process order storage request
	zap.S().Infof("Storage: Order '%s' (%s) getting stored", orderReq.Name, orderReq.ID)
	storageChannel <- orderReq
}

func internalProcess() {
	go func() {
		for {
			select {
			case orderReq, isOpen := <-storageChannel:
				if !isOpen {
					storageChannel = nil
					break
				}
				// Send order stored event
				zap.S().Infof("Storage: Order '%s'(%s) is stored at %s", orderReq.Name, orderReq.ID, time.Now())
			}

			if storageChannel == nil {
				break
			}
		}
	}()
}

func Stop() {
	close(storageChannel)
}
