package main

import (
	"math/rand"
	"time"
)

func NormalDist(stddev, mean float64, size int) []float64 {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	distribution := make([]float64, size)
	for i := 0; i < size; i++ {
		distribution[i] = r.NormFloat64()*stddev + mean
	}
	return distribution
}
