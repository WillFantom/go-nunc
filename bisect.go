package nunc

import (
	"sort"
)

type bisectFunc func(data []float64, value float64) (func(index int) bool, int)

var (
	bisectLeft bisectFunc = func(data []float64, value float64) (func(index int) bool, int) {
		return func(index int) bool {
			return data[index] >= value
		}, 0
	}
	bisectRight bisectFunc = func(data []float64, value float64) (func(index int) bool, int) {
		return func(index int) bool {
			return data[index] > value
		}, len(data)
	}
)

// bisect uses a bisect function to determine where in a given ordered float64
// slice a given quantile value should be inserted.
func bisect(f bisectFunc, data []float64, quantile float64) int {
	bf, def := f(data, quantile)
	index := sort.Search(len(data), bf)
	if index == -1 {
		return def
	}
	return index
}
