package model

type ShelfCapacity struct {
	Hot    int
	Cold   int
	Frozen int
	Any    int
}

func InitShelvesWithCapacity() ShelfCapacity {
	return ShelfCapacity{
		Hot:    10,
		Cold:   10,
		Frozen: 10,
		Any:    15,
	}
}

var ShelvesCapacity ShelfCapacity = InitShelvesWithCapacity()
