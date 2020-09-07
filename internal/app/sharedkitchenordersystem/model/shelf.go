package model

import (
	"time"
)

const HOT string = "hot"
const COLD string = "cold"
const FROZEN string = "frozen"
const OVERFLOW string = "overflow"

type ShelfCapacity struct {
	Hot    int
	Cold   int
	Frozen int
	Any    int
}

type ShelfItem struct {
	Order        Order
	CreatedTime  time.Time
	MaxLifeTimeS int64
}

func InitShelvesWithCapacity() ShelfCapacity {
	return ShelfCapacity{
		Hot:    1,
		Cold:   1,
		Frozen: 1,
		Any:    1,
	}
}

var ShelvesCapacity ShelfCapacity = InitShelvesWithCapacity()
