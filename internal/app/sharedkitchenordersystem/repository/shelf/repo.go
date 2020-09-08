package repo

import (
	"container/heap"
	"errors"
	"fmt"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"sync"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Value    model.ShelfItem // The value of the item; shelf item.
	Priority int64           // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

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

func (pq *PriorityQueue) Peek() interface{} {
	if pq.Len() > 0 {
		val := *pq
		return val[0]
	}
	return nil
}

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
	Push(model.ShelfItem)
	Pop() (model.ShelfItem, error)
	Size() int
	Peek() (model.ShelfItem, error)
	IsPresent(itemID string) bool
	GetRandomItem() (model.ShelfItem, error)
	Delete(string) error
}

type Shelf struct {
	sorter      PriorityQueue
	rack        map[string]*Item
	shelfLocker sync.Mutex
}

func (shelf *Shelf) Size() int {
	shelf.shelfLocker.Lock()
	defer shelf.shelfLocker.Unlock()

	return shelf.sorter.Len()
}

func (shelf *Shelf) Push(shelfItem model.ShelfItem) {
	shelf.shelfLocker.Lock()
	item := &Item{Value: shelfItem, Priority: shelfItem.MaxLifeTimeS}
	shelf.sorter.Push(item)
	shelf.rack[shelfItem.Order.ID] = item
	shelf.shelfLocker.Unlock()
}

func (shelf *Shelf) Pop() (model.ShelfItem, error) {
	shelf.shelfLocker.Lock()
	defer shelf.shelfLocker.Unlock()

	item := shelf.sorter.Pop()

	if item == nil {
		return (model.ShelfItem{}), errors.New("No items available to pop, shelf is empty!")
	}

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

var HotShelf IShelf
var ColdShelf IShelf
var FrozenShelf IShelf
var OverflowShelf map[string]IShelf

func Initialize() {
	HotShelf = &Shelf{
		sorter: make(PriorityQueue, 0),
		rack:   make(map[string]*Item),
	}

	ColdShelf = &Shelf{
		sorter: make(PriorityQueue, 0),
		rack:   make(map[string]*Item),
	}

	FrozenShelf = &Shelf{
		sorter: make(PriorityQueue, 0),
		rack:   make(map[string]*Item),
	}

	OverflowShelf = make(map[string]IShelf, 0)

	for _, temp := range []string{model.HOT, model.COLD, model.FROZEN} {
		OverflowShelf[temp] = &Shelf{
			sorter: make(PriorityQueue, 0),
			rack:   make(map[string]*Item),
		}
	}
}
