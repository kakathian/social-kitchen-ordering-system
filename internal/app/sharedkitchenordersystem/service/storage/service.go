package storage

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	repo "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/shelf"
	"time"

	"go.uber.org/zap"
)

// ProcessOrders .
var storageChannel chan model.ShelfItem = nil

func Start(noOfOrdersToRead int) {
	storageChannel = make(chan model.ShelfItem, noOfOrdersToRead)
	repo.Initialize()
	internalProcess()
}

func Process(shelfItem model.ShelfItem) {
	// process order storage request
	zap.S().Infof("Storage: Order '%s' (%s) getting stored", shelfItem.Order.Name, shelfItem.Order.ID)
	storageChannel <- shelfItem
}

func internalProcess() {
	go func() {
		for {
			select {
			case shelfItem, isOpen := <-storageChannel:
				if !isOpen {
					storageChannel = nil
					break
				}

				repo.ShelfLocker.Lock()
				arrangeItem(shelfItem)
				repo.ShelfLocker.Unlock()
				// Send order stored event
				zap.S().Infof("Storage: Order '%s'(%s) is stored at %s", shelfItem.Order.Name, shelfItem.Order.ID, time.Now())
				zap.S().Infof("Storage: Total number of items in shelf '%d' at %s", repo.Sorter.Len(), time.Now())
			}

			if storageChannel == nil {
				break
			}
		}
	}()
}

func arrangeItem(shelfItem model.ShelfItem) {
	item := repo.Sorter.Peek()
	for item != nil {
		sItem := (item.(*repo.Item)).Value
		if int64(time.Now().Sub(sItem.CreatedTime).Seconds())-sItem.MaxLifeTimeS >= 0 {
			repo.Sorter.Pop()
			zap.S().Infof("Storage: Total number of items in shelf '%d' at %s", repo.Sorter.Len(), time.Now())
			delete(repo.Rack, sItem.Order.ID)
		} else {
			break
		}

		item = repo.Sorter.Peek()
	}

	if repo.Sorter.Len() >= model.ShelvesCapacity.Hot {
		zap.S().Infof("Storage: Reached shelf capacity: Order '%s'(%s) is stored at %s", shelfItem.Order.Name, shelfItem.Order.ID, time.Now())
		return
	}

	sorterItem := &repo.Item{Value: shelfItem, Priority: shelfItem.MaxLifeTimeS}
	repo.Sorter.Push(sorterItem)
	repo.Rack[shelfItem.Order.ID] = sorterItem
}

func Stop() {
	close(storageChannel)
}
