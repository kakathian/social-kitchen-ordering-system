package dispatch

import (
	"math/rand"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	repo "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/shelf"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/storage"
	"strings"
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

				// Courier arrived randomly after this time
				// time.Sleep(time.Duration(rand.Intn(6-2+1)+2) * time.Second)
				time.Sleep(6 * time.Second)

				// Courier picking up the order
				var shelf *repo.Shelf

				if strings.EqualFold(orderReq.Temp, model.HOT) {
					shelf = &repo.HotShelf
				} else if strings.EqualFold(orderReq.Temp, model.COLD) {
					shelf = &repo.ColdShelf
				} else if strings.EqualFold(orderReq.Temp, model.FROZEN) {
					shelf = &repo.FrozenShelf
				}

				// If item not available in normal racks, check in overflow rack
				shelf.ShelfLocker.Lock()
				_, isPresent := shelf.Rack[orderReq.ID]
				shelf.ShelfLocker.Unlock()
				isOrderDispatched := false
				pickedUpShelfType := ""

				// Check if present in normal shelves
				if isPresent {
					shelf.ShelfLocker.Lock()
					shelf.Sorter.Delete(shelf.Rack[orderReq.ID])
					delete(shelf.Rack, orderReq.ID)
					shelf.ShelfLocker.Unlock()
					isOrderDispatched = true
					pickedUpShelfType = orderReq.Temp
					zap.S().Infof("Dispatch: Order '%s'(%s) removed from shelf '%s' by courier", orderReq.ID, orderReq.Name, orderReq.Temp)
					storage.NewSpaceAvailableChannel <- orderReq.Temp
					zap.S().Infof("Dispatch: New space available in shelf for '%s'", orderReq.Temp)
				} else {
					// Check if present in overflow shelf
					overflownShelf := repo.OverflowShelf[strings.ToLower(orderReq.Temp)]
					overflownShelf.ShelfLocker.Lock()
					if _, isPresent := overflownShelf.Rack[orderReq.ID]; isPresent {
						overflownShelf.Sorter.Delete(overflownShelf.Rack[orderReq.ID])
						delete(overflownShelf.Rack, orderReq.ID)
						isOrderDispatched = true
						pickedUpShelfType = model.OVERFLOW
					}

					overflownShelf.ShelfLocker.Unlock()
				}

				if isOrderDispatched {
					zap.S().Infof("Dispatch: Courier picked up Order '%s'(%s) from '%s' shelf ", orderReq.Name, orderReq.ID, pickedUpShelfType)
				} else {
					// TODO: Order could not be found, probably discarded - should be confirmed discarded/expired onlyif present in trash bin
					zap.S().Infof("Dispatch: Courier could not find the Order '%s'(%s) in shelves; probably 'discarded'", orderReq.Name, orderReq.ID)
				}
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
