package nunc

import (
	"fmt"
	"math"
	"sort"
)

// NuncProcessor is a data processor for determining changepoints in data feeds.
// Implementing the NUNC algorithm, a NUNC processor takes floting point values
// and determines if and when a change in the data's distribution has occurred.
type NuncProcessor struct {
	window *Window[float64]
}

// Cost is an attributed "cost" of a datapoint in the feed as determined by the
// NUNC processor. This in itself is just data, to decide if the point is or is
// not a changepoint, the value must be compared against some threshold.
type Cost struct {
	dataIndex uint64
	value     float64
}

// Value returns the associated "cost" of the datapoint as computed by the NUNC
// processor.
func (c Cost) Value() float64 {
	return c.value
}

// Index returns the datapoint index where the maximum cost occurred.
// Importantly, this is not the index as to when the point was detected, but
// when the datapoint index that the changepoint is estimated to have occurred
// on.
func (c Cost) Index() uint64 {
	return c.dataIndex
}

// New returns a new NUNC processor with the given window size. If the window
// size is not valid, an error is returned.
func New(windowSize int) (*NuncProcessor, error) {
	w, err := NewWindow[float64](windowSize)
	if err != nil {
		return nil, err
	}
	return &NuncProcessor{
		window: w,
	}, nil
}

// Process performs the NUNC logic on the datapoint presented as the value.
// Returned is the cost associated with the datapoint if one can be calculated.
// If a cost can not be computed, an error is returned. This can safely be used
// from multiple routines.
func (proc *NuncProcessor) Process(value float64, quantileCount int) (*Cost, error) {
	count, dataset := proc.window.PushGetFull(value)
	if dataset == nil {
		return nil, fmt.Errorf("window is not yet fully populated")
	}
	sorted := make([]float64, len(dataset))
	copy(sorted, dataset)
	sort.Float64s(sorted)

	// get quantiles
	q := quantiles(sorted, quantileCount)

	// get cdf values and full cost
	fullCDF := make([]float64, len(q))
	fullCost := 0.0
	for i := 0; i < len(q); i++ {
		fullCDF[i] = ecdf(sorted, q[i])
		fullCost += cdfCost(fullCDF[i], len(dataset))
	}

	// loop to calculate segment costs
	rightCDF := make([]float64, len(fullCDF))
	leftCDF := make([]float64, len(q))
	copy(rightCDF, fullCDF)
	maxCost := new(Cost)
	for i := 0; i < len(sorted); i++ {
		length := len(sorted) - i
		rightCDF = windowUpdate(dataset[i], length, q, rightCDF)
		length -= 1
		for j := 0; j < len(q); j++ {
			leftCDF[j] = ((fullCDF[j] * float64(len(sorted))) - (rightCDF[j] * float64(length))) / (float64(len(sorted) - length))
		}
		leftCost := 0.0
		for _, c := range leftCDF {
			leftCost += cdfCost(c, len(sorted)-length)
		}
		rightCost := 0.0
		for _, c := range rightCDF {
			rightCost += cdfCost(c, length)
		}
		cost := 2.0 * (leftCost + rightCost - fullCost)
		if cost > maxCost.value {
			maxCost.dataIndex = count - 1 - uint64(proc.window.Cap()) + uint64(i)
			maxCost.value = cost
		}
	}

	return maxCost, nil
}

func quantiles(data []float64, quantileCount int) []float64 {
	quantiles := make([]float64, quantileCount)
	c := math.Log(float64((2 * len(data)) - 1))
	for i := 0; i < quantileCount; i++ {
		pct := 1.0 / (1.0 + (2.0*(float64(len(data))-1.0))*math.Exp((-c/float64(quantileCount))*(2.0*float64(i)-1.0)))
		quantiles[i] = quantile(data, pct)
	}
	return quantiles
}

func quantile(data []float64, pct float64) float64 {
	index := float64(len(data)-1) * pct
	lower := data[int(math.Floor(index))]
	upper := data[int(math.Ceil(index))]
	if lower == upper {
		return lower
	}
	return lower + ((index - math.Floor(index)) * (upper - lower))
}

func ecdf(data []float64, quantile float64) float64 {
	left := bisect(bisectLeft, data, quantile)
	right := bisect(bisectRight, data, quantile)
	return (float64(left) + float64((right-left)/2)) / float64(len(data))
}

func empDist(datapoint float64, quantile float64) float64 {
	if datapoint < quantile {
		return 1.0
	} else if datapoint > quantile {
		return 0.0
	}
	return 0.5
}

func cdfCost(value float64, dataSize int) float64 {
	if value <= 0 || value >= 1 {
		return 0.0
	}
	conj := 1 - value
	return float64(dataSize) * ((value * math.Log(value)) - (conj * math.Log(conj)))
}

func windowUpdate(datapoint float64, length int, quantiles, cdf []float64) []float64 {
	for i := 0; i < len(quantiles); i++ {
		cdf[i] *= float64(length)
		cdf[i] -= empDist(datapoint, quantiles[i])
		cdf[i] /= float64(length - 1)
	}
	return cdf
}
