package dispatch

import (
	"math/rand"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	repo "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/shelf"
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
				rand.Seed(time.Now().UnixNano())

				// Courier arrived randomly fter this time
				time.Sleep(time.Duration(rand.Intn(6-2+1)+2) * time.Second)

				// Courier picking up the order
				repo.ShelfLocker.Lock()
				repo.Sorter.Delete(repo.Rack[orderReq.ID])
				delete(repo.Rack, orderReq.ID)
				repo.ShelfLocker.Unlock()
				zap.S().Infof("Dispatch: Courier picked up Order '%s'(%s) at %s", orderReq.Name, orderReq.ID, time.Now())
				zap.S().Infof("Dispatch: Total number of items in shelf-sorter (%d): rack (%d)", repo.Sorter.Len(), len(repo.Rack))
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
