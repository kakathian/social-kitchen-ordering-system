package model

type Order struct {
	Id string `json:"id"`

	Name string `json:"name"`

	Temp string `json:"temp"`

	// Shelf wait max duration (seconds)"decayRate": ​0.45​
	ShelfLife int32 `json:"shelfLife"`

	DecayRate float32 `json:"decayRate"`
}

type Orders struct {
	Orders []Order
}
