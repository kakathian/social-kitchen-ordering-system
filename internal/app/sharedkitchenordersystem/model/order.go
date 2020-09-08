package model

import (
	"time"
)

const ORDER_RECEIVED string = "received"
const ORDER_PROCESSED string = "processed"
const ORDER_PICKED string = "picked"
const ORDER_EXPIRED string = "expired"
const ORDER_EVICTED string = "evicted"

// Order .
type Order struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Temp string `json:"temp"`

	// Shelf wait max duration (seconds)"decayRate": ​0.45​
	ShelfLife int32 `json:"shelfLife"`

	DecayRate float32 `json:"decayRate"`
}

// Orders .
type Orders struct {
	Orders []Order
}

// OrderRequest .
type OrderRequest struct {
	Order Order
	Time  time.Time
}

type OrderStatus struct {
	Status  string
	OrderId string
}
