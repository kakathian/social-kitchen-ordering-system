package model

import (
	"time"
)

const HOT string = "hot"
const COLD string = "cold"
const FROZEN string = "frozen"
const OVERFLOW string = "overflow"

type ShelfItem struct {
	Order        Order
	CreatedTime  time.Time
	MaxLifeTimeS int64
}
