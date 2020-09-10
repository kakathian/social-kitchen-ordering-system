package storage

import (
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/model"
	repo "sharedkitchenordersystem/internal/app/sharedkitchenordersystem/repository/shelf"
	"sharedkitchenordersystem/internal/app/sharedkitchenordersystem/service/supervisor"
	"sharedkitchenordersystem/internal/pkg"
	"strings"
	"testing"
	"time"
)

func Test_storeItem(t *testing.T) {
	type args struct {
		shelfItem model.ShelfItem
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_storeItem_ShouldStore_HotShelfItem_Success",
			args: args{model.ShelfItem{Order: model.Order{ID: "1", Name: "chicken", DecayRate: 1, ShelfLife: 20, Temp: "hot"},
				MaxLifeTimeS: pkg.CalculateMaxAge(20, 1, 1), CreatedTime: time.Now()}},
		},
		{
			name: "Test_storeItem_ShouldStore_ColdShelfItem_Success",
			args: args{model.ShelfItem{Order: model.Order{ID: "2", Name: "juice", DecayRate: 1, ShelfLife: 10, Temp: "cold"},
				MaxLifeTimeS: pkg.CalculateMaxAge(10, 1, 1), CreatedTime: time.Now()}},
		},
		{
			name: "Test_storeItem_ShouldStore_FrozenShelfItem_Success",
			args: args{model.ShelfItem{Order: model.Order{ID: "3", Name: "ice cream", DecayRate: 1, ShelfLife: 15, Temp: "frozen"},
				MaxLifeTimeS: pkg.CalculateMaxAge(15, 1, 1), CreatedTime: time.Now()}},
		},
	}
	repo.Initialize()
	supervisor.Start(1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeItem(tt.args.shelfItem)
			if shelf, err := repo.ShelfFactory(tt.args.shelfItem.Order.Temp); err != nil || !shelf.IsPresent(tt.args.shelfItem.Order.ID) {
				t.Errorf("internalProcess(), order %v not stored", tt.args.shelfItem.Order.ID)
			}
		})
	}
}

func getShelf(temp string) repo.IShelf {
	s, _ := repo.ShelfFactory(temp)
	return s
}

func Test_checkAndRemoveOverflownExpiredOrders(t *testing.T) {
	repo.Initialize()
	supervisor.Start(1)

	type args struct {
		shelf     repo.IShelf
		shelfItem model.ShelfItem
	}
	tests := []struct {
		name          string
		args          args
		mustBeRemoved bool
	}{
		{
			name: "Test_checkAndRemoveOverflownExpiredOrders_ItemExpired_MustBeRemoved",
			args: args{
				shelf: getShelf(model.HOT),
				shelfItem: model.ShelfItem{Order: model.Order{
					ID: "3", Name: "chicken", DecayRate: 1, ShelfLife: 10, Temp: "hot"}, MaxLifeTimeS: 5, CreatedTime: time.Now().Add(-5 * time.Second)}},
			mustBeRemoved: true,
		},
		{
			name: "Test_checkAndRemoveOverflownExpiredOrders_ItemExpired_MustNotBeRemoved",
			args: args{
				shelf: getShelf(model.HOT),
				shelfItem: model.ShelfItem{Order: model.Order{
					ID: "3", Name: "chicken", DecayRate: 1, ShelfLife: 10, Temp: "hot"}, MaxLifeTimeS: 100, CreatedTime: time.Now().Add(time.Second)}},
			mustBeRemoved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.shelf.Push(tt.args.shelfItem)
			checkAndRemoveOverflownExpiredOrders(tt.args.shelf, tt.args.shelfItem)
			isRemoved := !tt.args.shelf.IsPresent(tt.args.shelfItem.Order.ID)
			if isRemoved != tt.mustBeRemoved {
				t.Errorf("checkAndRemoveOverflownExpiredOrders(), got isExpiredRemoved:%v, want %v ", isRemoved, tt.mustBeRemoved)
			}
		})
	}
}

func Test_removeOrders(t *testing.T) {
	repo.Initialize()
	supervisor.Start(1)

	type args struct {
		shelf     repo.IShelf
		shelfItem model.ShelfItem
	}
	tests := []struct {
		name          string
		args          args
		mustBeRemoved bool
	}{
		{
			name: "Test_removeOrders_ItemExpired_MustBeRemoved",
			args: args{
				shelf: getShelf(model.HOT),
				shelfItem: model.ShelfItem{Order: model.Order{
					ID: "3", Name: "chicken", DecayRate: 1, ShelfLife: 10, Temp: "hot"}, MaxLifeTimeS: 5, CreatedTime: time.Now().Add(-5 * time.Second)}},
			mustBeRemoved: true,
		},
		{
			name: "Test_removeOrders_ItemExpired_MustNotBeRemoved",
			args: args{
				shelf: getShelf(model.HOT),
				shelfItem: model.ShelfItem{Order: model.Order{
					ID: "1", Name: "chicken", DecayRate: 1, ShelfLife: 10, Temp: "hot"}, MaxLifeTimeS: 100, CreatedTime: time.Now().Add(time.Second)}},
			mustBeRemoved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.shelf.Push(tt.args.shelfItem)
			removeOrders(tt.args.shelf, tt.args.shelfItem)
			isRemoved := !tt.args.shelf.IsPresent(tt.args.shelfItem.Order.ID)
			if isRemoved != tt.mustBeRemoved {
				t.Errorf("removeOrders(), got isExpiredRemoved:%v, want %v ", isRemoved, tt.mustBeRemoved)
			}
		})
	}
}

func Test_onSpaceOverflownEventReceived(t *testing.T) {
	repo.Initialize()
	supervisor.Start(10)

	type args struct {
		shelf     repo.IShelf
		shelfItem model.ShelfItem
	}
	tests := []struct {
		name         string
		args         args
		mustBeStored bool
	}{
		{
			name: "Test_onSpaceOverflownEventReceived_ItemExpired_NotStored",
			args: args{
				shelf: getShelf(model.HOT),
				shelfItem: model.ShelfItem{Order: model.Order{
					ID: "10", Name: "chicken", DecayRate: 1, ShelfLife: 10, Temp: "hot"}, MaxLifeTimeS: pkg.CalculateMaxAge(10, 1, 1), CreatedTime: time.Now().Add(-5 * time.Second)}},
			mustBeStored: false,
		},
		{
			name: "Test_onSpaceOverflownEventReceived_ItemNotExpired_IsStored",
			args: args{
				shelf: getShelf(model.COLD),
				shelfItem: model.ShelfItem{Order: model.Order{
					ID: "20", Name: "ice cream", DecayRate: 1, ShelfLife: 100, Temp: "cold"}, MaxLifeTimeS: pkg.CalculateMaxAge(100, 0.2, 1), CreatedTime: time.Now()}},
			mustBeStored: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			onSpaceOverflownEventReceived(tt.args.shelfItem)
			overflownShelf := repo.OverflowShelf[strings.ToLower(tt.args.shelfItem.Order.Temp)]
			isStored := overflownShelf.IsPresent(strings.ToLower(tt.args.shelfItem.Order.ID))
			if isStored != tt.mustBeStored {
				t.Errorf("processSpaceOverflownEvents(), got %v, want %v ", isStored, tt.mustBeStored)
			}
		})
	}
}

func Test_onNewShelfSpaceAvailableReceived(t *testing.T) {
	repo.Initialize()
	supervisor.Start(10)

	type args struct {
		shelf     repo.IShelf
		shelfItem model.ShelfItem
	}
	tests := []struct {
		name                            string
		args                            args
		mustBeRemovedFromOverflownShelf bool
	}{
		{
			name: "Test_onNewShelfSpaceAvailableReceived_RemoveItem_SendToStore",
			args: args{
				shelf: getShelf(model.HOT),
				shelfItem: model.ShelfItem{Order: model.Order{
					ID: "10", Name: "chicken", DecayRate: 1, ShelfLife: 10, Temp: "hot"}, MaxLifeTimeS: pkg.CalculateMaxAge(10, 1, 1), CreatedTime: time.Now().Add(-5 * time.Second)}},
			mustBeRemovedFromOverflownShelf: true,
		},
		{
			name: "Test_onNewShelfSpaceAvailableReceived_RemoveItem_SendToStore",
			args: args{
				shelf: getShelf(model.COLD),
				shelfItem: model.ShelfItem{Order: model.Order{
					ID: "10", Name: "chicken", DecayRate: 0.5, ShelfLife: 100, Temp: "cold"}, MaxLifeTimeS: pkg.CalculateMaxAge(100, 1, 1), CreatedTime: time.Now()}},
			mustBeRemovedFromOverflownShelf: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overflownShelf := repo.OverflowShelf[strings.ToLower(tt.args.shelfItem.Order.Temp)]
			overflownShelf.Push(tt.args.shelfItem)
			onNewShelfSpaceAvailableReceived(tt.args.shelfItem.Order.Temp)
			isRemoved := !overflownShelf.IsPresent(strings.ToLower(tt.args.shelfItem.Order.ID))
			if isRemoved != tt.mustBeRemovedFromOverflownShelf {
				t.Errorf("onNewShelfSpaceAvailableReceive(), got %v, want %v ", isRemoved, tt.mustBeRemovedFromOverflownShelf)
			}
		})
	}
}
