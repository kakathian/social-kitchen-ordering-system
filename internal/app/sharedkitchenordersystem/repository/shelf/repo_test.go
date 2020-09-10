package repo

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"testing"
)

func TestPriorityQueuePush(t *testing.T) {
	//pq := make(PriorityQueue, 0)
	//heap.Init(&pq)

	// shelfItem1 := &Item{Value: model.ShelfItem{Order: model.Order{ID: "1", Name: "orange"}, MaxLifeTimeS: 30}, Priority: 30}
	// shelfItem2 := &Item{Value: model.ShelfItem{Order: model.Order{ID: "2", Name: "apple"}, MaxLifeTimeS: 10}, Priority: 10}
	// shelfItem3 := &Item{Value: model.ShelfItem{Order: model.Order{ID: "3", Name: "mango"}, MaxLifeTimeS: 20}, Priority: 20}

	// heap.Push(&pq, shelfItem1)
	// heap.Push(&pq, shelfItem2)
	// heap.Push(&pq, shelfItem3)
	// for pq.Len() > 0 {
	// 	item := heap.Pop(&pq).(*Item)
	// 	t.Errorf("%.2d:%s ", item.Priority, item.Value.Order.Name)
	// }
	shelfItem1 := model.ShelfItem{Order: model.Order{ID: "1", Name: "juice"}, MaxLifeTimeS: 10}
	shelfItem2 := model.ShelfItem{Order: model.Order{ID: "2", Name: "icecream"}, MaxLifeTimeS: 20}
	shelfItem3 := model.ShelfItem{Order: model.Order{ID: "3", Name: "chicken"}, MaxLifeTimeS: 30}
	shelfItem4 := model.ShelfItem{Order: model.Order{ID: "4", Name: "egg sandwich"}, MaxLifeTimeS: 1}

	shelf := &Shelf{
		sorter: make(PriorityQueue, 0),
		rack:   make(map[string]*Item),
	}
	shelf.Init()

	shelf.Push(shelfItem2)
	shelf.Push(shelfItem1)
	shelf.Push(shelfItem3)

	shelfItem, _ := shelf.Pop()

	if shelfItem.Order.Name != shelfItem1.Order.Name {
		t.Errorf("PriorityQueue sorting incorrect, got order with priority: %d, want: %d", shelfItem.MaxLifeTimeS, shelfItem1.MaxLifeTimeS)
	}

	shelf.Push(shelfItem4)

	shelfItem, _ = shelf.Peek()

	if shelfItem.Order.Name != shelfItem4.Order.Name {
		t.Errorf("PriorityQueue aaqaa sorting incorrect, got order with priority: %d, want: %d", shelfItem.MaxLifeTimeS, shelfItem4.MaxLifeTimeS)
	}
}
