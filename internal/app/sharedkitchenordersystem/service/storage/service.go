package storage

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	repo "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/shelf"
	"sharedkitchenordersystem/internal/pkg"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ProcessOrders .
var storageChannel chan model.ShelfItem = nil
var NewSpaceAvailableChannel chan string = nil
var OverflownChannel chan model.ShelfItem = nil

func Start(noOfOrdersToRead int) {
	storageChannel = make(chan model.ShelfItem, noOfOrdersToRead)
	NewSpaceAvailableChannel = make(chan string, noOfOrdersToRead)
	OverflownChannel = make(chan model.ShelfItem, noOfOrdersToRead)
	repo.Initialize()
	internalProcess()
}

func Process(shelfItem model.ShelfItem) {
	// process order storage request
	zap.S().Infof("Storage: Order '%s' (%s) getting stored", shelfItem.Order.Name, shelfItem.Order.ID)
	storageChannel <- shelfItem
}

func internalProcess() {
	go processSpaceOverflownEvents()
	go processNewShelfSpaceAvailable()
	go collectTempControlledShelvesExpiredOrders()
	go collectOverflownShelveExpiredOrders()

	go func() {
		for {
			select {
			case shelfItem, isOpen := <-storageChannel:
				if !isOpen {
					storageChannel = nil
					break
				}
				storeItem(shelfItem)
				// Send order stored event
				zap.S().Infof("Storage: Order '%s'(%s) is stored at %s", shelfItem.Order.Name, shelfItem.Order.ID, time.Now())
			}

			if storageChannel == nil {
				break
			}
		}
	}()
}

func storeItem(shelfItem model.ShelfItem) {
	var shelf *repo.Shelf
	var capacity int
	var currentLen int

	if strings.EqualFold(shelfItem.Order.Temp, model.HOT) {
		shelf = &repo.HotShelf
		capacity = model.ShelvesCapacity.Hot
	} else if strings.EqualFold(shelfItem.Order.Temp, model.COLD) {
		shelf = &repo.ColdShelf
		capacity = model.ShelvesCapacity.Cold
	} else if strings.EqualFold(shelfItem.Order.Temp, model.FROZEN) {
		shelf = &repo.FrozenShelf
		capacity = model.ShelvesCapacity.Frozen
	}

	shelf.ShelfLocker.Lock()
	currentLen = shelf.Sorter.Len()
	shelf.ShelfLocker.Unlock()

	if currentLen >= capacity {
		zap.S().Infof("Storage: Reached %s shelf capacity: Raise overflown event for Order '%s'(%s)", shelfItem.Order.Temp, shelfItem.Order.Name, shelfItem.Order.ID)

		// Raise overflow event
		OverflownChannel <- shelfItem
		return
	}

	sorterItem := &repo.Item{Value: shelfItem, Priority: shelfItem.MaxLifeTimeS}

	shelf.ShelfLocker.Lock()

	shelf.Sorter.Push(sorterItem)
	shelf.Rack[shelfItem.Order.ID] = sorterItem

	shelf.ShelfLocker.Unlock()
}

// TODO: Refactor - reuse the other remove method
func collectOverflownShelveExpiredOrders() {
	for {
		// Remove overflown shelf expired orders
		for _, overflowCompartment := range repo.OverflowShelf {
			overflowCompartment.ShelfLocker.Lock()
			checkAndRemoveOverflownExpiredOrders(overflowCompartment, repo.HotShelf.Sorter.Peek())
			overflowCompartment.ShelfLocker.Unlock()
		}
		time.Sleep(time.Second)
	}
}

func checkAndRemoveOverflownExpiredOrders(shelf *repo.Shelf, item interface{}) {
	if item == nil {
		return
	}

	for item != nil {
		sItem := (item.(*repo.Item)).Value
		if int64(time.Now().Sub(sItem.CreatedTime).Seconds())-sItem.MaxLifeTimeS >= 0 {
			shelf.Sorter.Pop()
			zap.S().Infof("Storage: Total number of items in shelf '%d' at %s", shelf.Sorter.Len(), time.Now())
			delete(shelf.Rack, sItem.Order.ID)
		} else {
			break
		}

		item = shelf.Sorter.Peek()
	}
}

func collectTempControlledShelvesExpiredOrders() {
	for {
		// Remove expired orders

		// HotShelf
		repo.HotShelf.ShelfLocker.Lock()
		hotShelftem := repo.HotShelf.Sorter.Peek()
		removeOrders(&repo.HotShelf, hotShelftem)
		repo.HotShelf.ShelfLocker.Unlock()
		time.Sleep(time.Second)

		// ColdShelf
		repo.ColdShelf.ShelfLocker.Lock()
		coldShelfitem := repo.ColdShelf.Sorter.Peek()
		removeOrders(&repo.ColdShelf, coldShelfitem)
		repo.ColdShelf.ShelfLocker.Unlock()
		time.Sleep(time.Second)

		// FrozenShelf
		repo.FrozenShelf.ShelfLocker.Lock()
		frozenShelfitem := repo.FrozenShelf.Sorter.Peek()
		removeOrders(&repo.FrozenShelf, frozenShelfitem)
		repo.FrozenShelf.ShelfLocker.Unlock()
		time.Sleep(time.Second)
	}
}

func removeOrders(shelf *repo.Shelf, item interface{}) {
	if item == nil {
		return
	}

	for item != nil {
		sItem := (item.(*repo.Item)).Value
		if int64(time.Now().Sub(sItem.CreatedTime).Seconds())-sItem.MaxLifeTimeS >= 0 {
			shelf.Sorter.Pop()
			zap.S().Infof("Storage: Order '%s'(%s) expired and removed", sItem.Order.ID, sItem.Order.Name)
			delete(shelf.Rack, sItem.Order.ID)

			// Fire event - SpaceAvailable
			NewSpaceAvailableChannel <- sItem.Order.Temp
			zap.S().Infof("Storage: New space available in shelf for '%s' at %s", sItem.Order.Temp, time.Now())
		} else {
			break
		}

		item = shelf.Sorter.Peek()
	}
}

// On SpaceOverflown event received, overflow shelf stores the overflown shelf item. If enough space is
// not available on overflow shelf, it will remove a random shelf item and stores the incoming shelf item
func processSpaceOverflownEvents() {
	for {
		select {
		case overflownShelfItem, isOpen := <-OverflownChannel:
			if !isOpen {
				OverflownChannel = nil
				break
			}

			zap.S().Infof("Storage: Overflow shelf received Order '%s'(%s) to store in overflow shelf", overflownShelfItem.Order.Name, overflownShelfItem.Order.ID)

			overflownShelf := repo.OverflowShelf[strings.ToLower(overflownShelfItem.Order.Temp)]
			var overflowShelfCurrentSize int = 0
			// Get size of all compartments together (total size is overflow shelf size)
			for _, compartment := range repo.OverflowShelf {
				compartment.ShelfLocker.Lock()
				overflowShelfCurrentSize += compartment.Sorter.Len()
				compartment.ShelfLocker.Unlock()
			}

			// Calculate max order age for overflow shelf
			maxAgeForOverflowShelf := pkg.CalculateMaxAge(overflownShelfItem.Order.ShelfLife, overflownShelfItem.Order.DecayRate, 2)
			currentOrderAge := int64(time.Now().Sub(overflownShelfItem.CreatedTime).Seconds())

			// Check if the order is not expired, if so discard it or else store
			if currentOrderAge >= maxAgeForOverflowShelf {
				zap.S().Infof("Storage: Overflow shelf marked order '%s'(%s) as trash because it is expired. Expected below %d(s) but was %d(s)", overflownShelfItem.Order.Name, overflownShelfItem.Order.ID, maxAgeForOverflowShelf, currentOrderAge)
				continue
			}

			// Check if overflow reached its max max capacity. If so, remove a random order and make some space available for incoming item
			if overflowShelfCurrentSize >= model.ShelvesCapacity.Any {
				zap.S().Infof("Storage: Overflow shelf reached its max size, removing random shelf item")
				overflownShelf.ShelfLocker.Lock()

				// Get random map key - go does not gurantee order of a map, so the order itself is random by default
				var randomOrderID string = ""

				for id, _ := range overflownShelf.Rack {
					randomOrderID = id
					break
				}

				if randomOrderID != "" {
					randomItem := overflownShelf.Rack[randomOrderID]
					randomOrderName := randomItem.Value.Order.Name

					overflownShelf.Sorter.Delete(randomItem)
					delete(overflownShelf.Rack, randomOrderID)
					zap.S().Infof("Storage: Overflow shelf removed random element: Order '%s'(%s)", randomOrderID, randomOrderName)

					randomItem = nil
					overflownShelf.ShelfLocker.Unlock()
				}
			}

			item := &repo.Item{
				Value:    overflownShelfItem,
				Priority: maxAgeForOverflowShelf - currentOrderAge,
			}

			overflownShelf.ShelfLocker.Lock()
			overflownShelf.Sorter.Push(item)
			overflownShelf.Rack[overflownShelfItem.Order.ID] = item

			overflownShelf.ShelfLocker.Unlock()

		}
		if OverflownChannel == nil {
			break
		}
	}
}

func processNewShelfSpaceAvailable() {
	for {
		select {
		case newShelfSpaceTempType, isOpen := <-NewSpaceAvailableChannel:
			if !isOpen {
				NewSpaceAvailableChannel = nil
				break
			}
			// Send order stored event
			zap.S().Infof("Storage: Overflow cabin received new shelf space available for %s temp", newShelfSpaceTempType)

			// On new shelf space available, promote an item from overflow shelf to corresponding shelf with that temperature
			var shelf *repo.Shelf = repo.OverflowShelf[strings.ToLower(newShelfSpaceTempType)]
			shelf.ShelfLocker.Lock()
			item := shelf.Sorter.Pop()

			if item != nil {
				sItem := (item.(*repo.Item)).Value
				delete(shelf.Rack, sItem.Order.ID)
				// TODO: Read constant factor for overflow shelf
				// Check if this item is already expired before moving to main shelf
				/*maxAge := pkg.CalculateMaxAge(sItem.Order.ShelfLife, sItem.Order.DecayRate, 1)
				if int64(time.Now().Sub(sItem.CreatedTime).Seconds())-maxAge >= 0 {
					zap.S().Infof("Storage: Order '%s' (%s) expired before moving to main shelf from overflow cabin", sItem.Order.Name, sItem.Order.ID)
					shelf.Sorter.Pop()
					delete(shelf.Rack, sItem.Order.ID)
					return
				}*/

				// Send StoreOrder event
				zap.S().Infof("Storage: Order '%s' (%s) removed from overflow and sent to store on normal temp shelf", sItem.Order.Name, sItem.Order.ID)
				Process(sItem)
				zap.S().Infof("Storage: Total number of items in shelf '%d' at %s", shelf.Sorter.Len(), time.Now())
			}

			shelf.ShelfLocker.Unlock()
		}

		if NewSpaceAvailableChannel == nil {
			break
		}
	}
}

func Stop() {
	close(storageChannel)
}
