package storage

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	repo "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/shelf"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/supervisor"
	"sharedkitchenordersystem/internal/pkg"
	"strings"
	"time"

	"go.uber.org/zap"
)

func Start(noOfOrdersToRead int) {
	repo.Initialize()
	internalProcess()
}

func internalProcess() {
	go processSpaceOverflownEvents()
	go processNewShelfSpaceAvailable()
	go collectTempControlledShelvesExpiredOrders()
	go collectOverflownShelveExpiredOrders()

	go func() {
		for {
			select {
			case shelfItem, isOpen := <-supervisor.StorageChannel:
				if !isOpen {
					supervisor.StorageChannel = nil
					break
				}
				zap.S().Infof("Storage: Order '%s' (%s) getting stored", shelfItem.Order.Name, shelfItem.Order.ID)

				storeItem(shelfItem)
				// Send order stored event
				zap.S().Infof("Storage: Order '%s'(%s) is stored at %s", shelfItem.Order.Name, shelfItem.Order.ID, time.Now())
			}

			if supervisor.StorageChannel == nil {
				break
			}
		}
	}()
}

func storeItem(shelfItem model.ShelfItem) {
	var shelf repo.IShelf
	var capacity int
	var currentLen int

	if strings.EqualFold(shelfItem.Order.Temp, model.HOT) {
		shelf = repo.HotShelf
		capacity = model.ShelvesCapacity.Hot
	} else if strings.EqualFold(shelfItem.Order.Temp, model.COLD) {
		shelf = repo.ColdShelf
		capacity = model.ShelvesCapacity.Cold
	} else if strings.EqualFold(shelfItem.Order.Temp, model.FROZEN) {
		shelf = repo.FrozenShelf
		capacity = model.ShelvesCapacity.Frozen
	}

	currentLen = shelf.Size()

	if currentLen >= capacity {
		zap.S().Infof("Storage: Reached %s shelf capacity: Raise overflown event for Order '%s'(%s)", shelfItem.Order.Temp, shelfItem.Order.Name, shelfItem.Order.ID)

		// Raise overflow event
		supervisor.OverflownChannel <- shelfItem
		return
	}

	shelf.Push(shelfItem)
}

// TODO: Refactor - reuse the other remove method
func collectOverflownShelveExpiredOrders() {
	for {
		// Remove overflown shelf expired orders
		for _, overflowCompartment := range repo.OverflowShelf {
			item, err := overflowCompartment.Peek()

			if err != nil {
				continue
			}

			checkAndRemoveOverflownExpiredOrders(overflowCompartment, item)
		}
		time.Sleep(time.Second)
	}
}

func checkAndRemoveOverflownExpiredOrders(shelf repo.IShelf, shelfItem model.ShelfItem) {
	if shelfItem == (model.ShelfItem{}) {
		return
	}

	var err error = nil
	for shelfItem != (model.ShelfItem{}) && err == nil {
		currAge := int64(time.Now().Sub(shelfItem.CreatedTime).Seconds())
		if currAge-shelfItem.MaxLifeTimeS >= 0 {
			shelf.Pop()

			// Send OrderStatus event
			supervisor.SupervisorChannel <- model.OrderStatus{OrderId: shelfItem.Order.ID, Status: model.ORDER_EXPIRED}

			zap.S().Infof("Storage: Order '%s'(%s) expired and removed; current age %s(s), max allowed age %s(s)", shelfItem.Order.ID, shelfItem.Order.Name, currAge, shelfItem.MaxLifeTimeS)
			zap.S().Infof("Storage: Total number of items in overflow shelf '%d' at %s", shelf.Size(), time.Now())
		} else {
			break
		}

		shelfItem, err = shelf.Peek()
	}
}

func collectTempControlledShelvesExpiredOrders() {
	for {

		// Remove expired orders
		// HotShelf
		hotShelftem, err := repo.HotShelf.Peek()
		if err == nil {
			removeOrders(repo.HotShelf, hotShelftem)
		}

		// ColdShelf
		coldShelfitem, err := repo.ColdShelf.Peek()
		if err == nil {
			removeOrders(repo.ColdShelf, coldShelfitem)
		}

		// FrozenShelf
		frozenShelfitem, err := repo.FrozenShelf.Peek()
		if err == nil {
			removeOrders(repo.FrozenShelf, frozenShelfitem)
		}

		time.Sleep(time.Second)
	}
}

// removeOrders Removes the order with lowest priority which is available at root of priorityqueue (priority - order age)
// The order which ages soon or already aged would be at top of the tree
func removeOrders(shelf repo.IShelf, shelfItem model.ShelfItem) {
	if shelfItem == (model.ShelfItem{}) {
		return
	}

	var err error = nil

	for shelfItem != (model.ShelfItem{}) && err == nil {
		currAge := int64(time.Now().Sub(shelfItem.CreatedTime).Seconds())
		if currAge-shelfItem.MaxLifeTimeS >= 0 {
			shelf.Pop()
			zap.S().Infof("Storage: Order '%s'(%s) expired and removed; current age %d(s), max allowed age %d(s)", shelfItem.Order.ID, shelfItem.Order.Name, currAge, shelfItem.MaxLifeTimeS)

			// Send OrderStatus event
			supervisor.SupervisorChannel <- model.OrderStatus{OrderId: shelfItem.Order.ID, Status: model.ORDER_EXPIRED}

			// Fire event - NewSpaceAvailable
			supervisor.NewSpaceAvailableChannel <- shelfItem.Order.Temp
			zap.S().Infof("Storage: New space available in shelf for '%s' at %s", shelfItem.Order.Temp, time.Now())
		} else {
			break
		}

		shelfItem, err = shelf.Peek()
	}
}

// On SpaceOverflown event received, overflow shelf stores the overflown shelf item. If enough space is
// not available on overflow shelf, it will remove a random shelf item and stores the incoming shelf item
func processSpaceOverflownEvents() {
	for {
		select {
		case overflownShelfItem, isOpen := <-supervisor.OverflownChannel:
			if !isOpen {
				supervisor.OverflownChannel = nil
				break
			}

			zap.S().Infof("Storage: Overflow shelf received Order '%s'(%s) to store in overflow shelf", overflownShelfItem.Order.Name, overflownShelfItem.Order.ID)

			overflownShelf := repo.OverflowShelf[strings.ToLower(overflownShelfItem.Order.Temp)]
			var overflowShelfCurrentSize int = 0
			// Get size of all compartments together (total size is overflow shelf size)
			for _, compartment := range repo.OverflowShelf {
				overflowShelfCurrentSize += compartment.Size()
			}

			// Calculate max order age for overflow shelf
			maxAgeForOverflowShelf := pkg.CalculateMaxAge(overflownShelfItem.Order.ShelfLife, overflownShelfItem.Order.DecayRate, 2)
			currentOrderAge := int64(time.Now().Sub(overflownShelfItem.CreatedTime).Seconds())

			// Check if the order is not expired, if so discard it or else store
			if currentOrderAge >= maxAgeForOverflowShelf {
				// Send OrderStatus event
				supervisor.SupervisorChannel <- model.OrderStatus{OrderId: overflownShelfItem.Order.ID, Status: model.ORDER_EXPIRED}

				zap.S().Infof("Storage: Overflow shelf marked order '%s'(%s) as trash because it is expired. Expected below %d(s) but was %d(s)", overflownShelfItem.Order.Name, overflownShelfItem.Order.ID, maxAgeForOverflowShelf, currentOrderAge)
				continue
			}

			// Check if overflow reached its max max capacity. If so, remove a random order and make some space available for incoming item
			if overflowShelfCurrentSize >= model.ShelvesCapacity.Any {
				zap.S().Infof("Storage: Overflow shelf reached its max size, removing random shelf item")

				randomItem, err := overflownShelf.GetRandomItem()

				if err == nil {
					overflownShelf.Delete(randomItem.Order.ID)
					zap.S().Infof("Storage: Overflow shelf removed random element: Order '%s'(%s)", randomItem.Order.ID, randomItem.Order.Name)

					// Send OrderStatus event
					supervisor.SupervisorChannel <- model.OrderStatus{OrderId: randomItem.Order.ID, Status: model.ORDER_EVICTED}
				}
			}

			// Update max age for overflown shelf
			overflownShelfItem.MaxLifeTimeS = maxAgeForOverflowShelf - currentOrderAge
			overflownShelf.Push(overflownShelfItem)
		}
		if supervisor.OverflownChannel == nil {
			break
		}
	}
}

func processNewShelfSpaceAvailable() {
	for {
		select {
		case newShelfSpaceTempType, more := <-supervisor.NewSpaceAvailableChannel:
			if !more {
				supervisor.NewSpaceAvailableChannel = nil
				break
			}
			// Send order stored event
			zap.S().Infof("Storage: Overflow cabin received new shelf space available for %s temp", newShelfSpaceTempType)

			// On new shelf space available, promote an item from overflow shelf to corresponding shelf with that temperature
			var shelf repo.IShelf = repo.OverflowShelf[strings.ToLower(newShelfSpaceTempType)]
			item, err := shelf.Pop()

			if err == nil {
				// Check if this item is already expired before moving to main shelf
				maxAgeForNormalShelves := pkg.CalculateMaxAge(item.Order.ShelfLife, item.Order.DecayRate, 1)
				currentAge := int64(time.Now().Sub(item.CreatedTime).Seconds())
				if currentAge-maxAgeForNormalShelves >= 0 {
					zap.S().Infof("Storage: Order '%s' (%s) expired before moving to main shelf from overflow cabin", item.Order.Name, item.Order.ID)
					// Send OrderStatus event
					supervisor.SupervisorChannel <- model.OrderStatus{OrderId: item.Order.ID, Status: model.ORDER_EVICTED}
					return
				}

				// Send StoreOrder event
				zap.S().Infof("Storage: Order '%s' (%s) removed from overflow and sent to store on normal temp shelf", item.Order.Name, item.Order.ID)
				item.MaxLifeTimeS = maxAgeForNormalShelves - currentAge
				supervisor.StorageChannel <- item
				zap.S().Infof("Storage: Total number of items in shelf '%d' at %s", shelf.Size(), time.Now())
			}
		}

		if supervisor.NewSpaceAvailableChannel == nil {
			break
		}
	}
}
