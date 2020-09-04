package repo

import (
	"container/heap"
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

func (pq *PriorityQueue) Delete(item *Item) {
	heap.Remove(pq, (*item).index)
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value model.ShelfItem, priority int64) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.index)
}

var Sorter PriorityQueue
var Rack map[string]*Item
var ShelfLocker sync.Mutex

func Initialize() {
	Sorter = make(PriorityQueue, 0)
	Rack = make(map[string]*Item)
}
