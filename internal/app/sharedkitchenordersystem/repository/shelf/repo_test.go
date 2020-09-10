package repo

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	"testing"
)

func TestPriorityQueuePush(t *testing.T) {
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

	// Pop
	shelfItem, _ := shelf.Pop()

	if shelfItem.Order.Name != shelfItem1.Order.Name {
		t.Errorf("PriorityQueue Pop incorrect, got order with priority: %d, want: %d", shelfItem.MaxLifeTimeS, shelfItem1.MaxLifeTimeS)
	}

	shelf.Push(shelfItem4)

	// Peek
	shelfItem, _ = shelf.Peek()

	if shelfItem.Order.Name != shelfItem4.Order.Name {
		t.Errorf("PriorityQueue Peek incorrect, got order with priority: %d, want: %d", shelfItem.MaxLifeTimeS, shelfItem4.MaxLifeTimeS)
	}

	// Get (Remove)
	shelf.Delete("2")

	if shelf.IsPresent("2") {
		t.Errorf("PriorityQueue delete and find incorrect, got order %s present, want: deleted", "2")
	}

	// Size
	if shelf.Size() != 2 {
		t.Errorf("PriorityQueue size incorrect, got %d , want: %d", shelf.Size(), 2)
	}

	// Get random item
	item, err := shelf.GetRandomItem()
	if err != nil {
		t.Errorf("PriorityQueue GetRandomItem incorrect, got error '%s' , want: no error", err)
	}

	if !shelf.IsPresent(item.Order.ID) {
		t.Errorf("PriorityQueue GetRandomItem  incorrect, got order %s not present, want: present", item.Order.ID)
	}
}
