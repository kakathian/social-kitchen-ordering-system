package repo

import (
	"container/heap"
	"errors"
	"fmt"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"sync"
)

// An Item is managed in priority queue
type Item struct {
	Value    model.ShelfItem // The value of the item; shelf item.
	Priority int64           // The priority of the item in the queue.
	index    int             // The index of the item in the heap.
}

// PriorityQueue implements heap.Interface and holds Items
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

// Less helps decide which item needs to Pop; we want the the item with the lowest priority
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push stores item in the priorityqueue
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

// Pop removes the item with lowest priority
func (pq *PriorityQueue) Pop() interface{} {

	old := *pq
	n := len(old)
	if n == 0 {
		return nil
	}
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// Peek peeks the item with the lowest priority
func (pq *PriorityQueue) Peek() interface{} {
	if pq.Len() > 0 {
		val := *pq
		return val[0]
	}
	return nil
}

// delete deletes a given item
func (pq *PriorityQueue) delete(item *Item) {
	heap.Remove(pq, (*item).index)
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value model.ShelfItem, priority int64) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.index)
}

type IShelf interface {
	// Init initializes the priority queue
	Init()

	// Push pushes and item into th priority queue
	Push(model.ShelfItem)

	// Pop removes an item with the lowest priority
	Pop() (model.ShelfItem, error)

	// Size gives the number of items present
	Size() int

	// Peek peeks the item with the lowest priority
	Peek() (model.ShelfItem, error)

	// IsPresent checks if an item is present
	IsPresent(itemID string) bool

	// GetRandomItem gets a randoim item and removes it if present
	GetRandomItem() (model.ShelfItem, error)

	// Delete removes an item from the shelf
	Delete(string) error

	// MaxCapacity gives the max number of items the shelf can hold
	MaxCapacity() int
}

type Shelf struct {
	sorter      PriorityQueue
	rack        map[string]*Item
	shelfLocker sync.Mutex
	maxCapacity int
}

func (shelf *Shelf) Init() {
	heap.Init(&shelf.sorter)
}

func (shelf *Shelf) Size() int {
	shelf.shelfLocker.Lock()
	defer shelf.shelfLocker.Unlock()

	return shelf.sorter.Len()
}

func (shelf *Shelf) Push(shelfItem model.ShelfItem) {
	shelf.shelfLocker.Lock()
	item := &Item{Value: shelfItem, Priority: shelfItem.MaxLifeTimeS}
	heap.Push(&shelf.sorter, item)
	shelf.rack[shelfItem.Order.ID] = item
	shelf.shelfLocker.Unlock()
}

func (shelf *Shelf) Pop() (model.ShelfItem, error) {
	shelf.shelfLocker.Lock()
	defer shelf.shelfLocker.Unlock()

	if shelf.sorter.Len() == 0 {
		return (model.ShelfItem{}), errors.New("No items available to pop, shelf is empty!")
	}

	item := heap.Pop(&shelf.sorter)

	shelfItem := (item.(*Item)).Value
	delete(shelf.rack, shelfItem.Order.ID)
	return shelfItem, nil
}

func (shelf *Shelf) Peek() (model.ShelfItem, error) {
	shelf.shelfLocker.Lock()
	defer shelf.shelfLocker.Unlock()

	if shelf.sorter.Len() == 0 {
		return (model.ShelfItem{}), errors.New("Could not peek item from shelf, because it is empty")
	}

	item := shelf.sorter.Peek()
	shelfItem := (item.(*Item)).Value
	return shelfItem, nil
}

func (shelf *Shelf) IsPresent(itemID string) bool {
	shelf.shelfLocker.Lock()
	defer shelf.shelfLocker.Unlock()
	_, isPresent := shelf.rack[itemID]
	return isPresent
}

func (shelf *Shelf) GetRandomItem() (model.ShelfItem, error) {
	shelf.shelfLocker.Lock()
	defer shelf.shelfLocker.Unlock()

	// Get random item - go does not gurantee order of a map, so the order itself is random by default
	for id, _ := range shelf.rack {
		randomItem := shelf.rack[id]
		item := randomItem.Value
		return item, nil
	}

	return (model.ShelfItem{}), errors.New("Shelf is empty!")
}

func (shelf *Shelf) Delete(shelfItemID string) error {
	shelf.shelfLocker.Lock()
	defer shelf.shelfLocker.Unlock()
	var item *Item
	var isPresent bool
	if item, isPresent = shelf.rack[shelfItemID]; !isPresent {
		return errors.New(fmt.Sprintf("Storage: Order %s not present", shelfItemID))
	}

	shelf.sorter.delete(item)
	delete(shelf.rack, shelfItemID)

	return nil
}

func (shelf *Shelf) MaxCapacity() int {
	return shelf.maxCapacity
}

var hotShelf IShelf
var coldShelf IShelf
var frozenShelf IShelf
var OverflowShelf map[string]IShelf

var shelves map[string]IShelf
var ShelfTemperatures []string

var ShelvesCapacity map[string]int

func Initialize() {
	ShelvesCapacity = map[string]int{
		model.HOT:      10,
		model.COLD:     10,
		model.FROZEN:   10,
		model.OVERFLOW: 15,
	}

	ShelfTemperatures := []string{model.HOT, model.COLD, model.FROZEN}
	shelves = make(map[string]IShelf)

	for _, shelfType := range ShelfTemperatures {
		shelves[shelfType] = &Shelf{
			sorter:      make(PriorityQueue, 0),
			rack:        make(map[string]*Item),
			maxCapacity: ShelvesCapacity[shelfType],
		}
		shelves[shelfType].Init()
	}

	OverflowShelf = make(map[string]IShelf, 0)

	for _, temp := range ShelfTemperatures {
		OverflowShelf[temp] = &Shelf{
			sorter: make(PriorityQueue, 0),
			rack:   make(map[string]*Item),
		}

		OverflowShelf[temp].Init()
	}
}

func ShelfFactory(shelfTemperature string) (IShelf, error) {
	if shelf, isPresent := shelves[shelfTemperature]; isPresent {
		return shelf, nil
	}

	return nil, errors.New(fmt.Sprintf("Invalid shelfTemperature '%s' ", shelfTemperature))
}
