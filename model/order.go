package model

type Order struct {
	id string

	name string

	temp string

	// Shelf wait max duration (seconds)"decayRate": ​0.45​
	shelfLife int32

	decayRate float32
}

type Orders struct {
	Orders []Order
}
