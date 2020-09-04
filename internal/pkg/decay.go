package pkg

func CalculateMaxAge(shelfLife int32, decayRate float32, factor int8) int64 {
	return int64((float32(shelfLife)/(float32(1)+decayRate*float32(factor)) + 0.5))
}
