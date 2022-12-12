package nunc

import (
	"fmt"
	"math"
)

type Threshold interface {
	// Changepoint recives the maximum cost as computed by the data process and in
	// turn determines if the datapoint represents a changepoint in the
	// distribution. If a changepoint can not be computed for any reason, such as
	// an insufficiently populated window, an error is returned.
	Changepoint(cost float64) (bool, error)

	Value() float64
}

// ThresholdStatic compares computed costs to a constant value. If the cost is
// greater than the defined value, it is considered to be a changepoint.
type ThresholdStatic struct {
	value float64
}

func (ts ThresholdStatic) Changepoint(cost float64) (bool, error) {
	return (cost > ts.value), nil
}

func (ts ThresholdStatic) Value() float64 {
	return ts.value
}

// NewThresholdStatic creates a threshold that uses a single constant value as
// the the changepoint cost threshold.
func NewThresholdStatic(value float64) Threshold {
	return ThresholdStatic{value: value}
}

// ThresholdEstimate determines a threshold value estimated to be suitable for
// the given parameters. The logic of this estimation focuses on the chance of a
// false changepoint being detected in a given number of datapoints. For
// example, the estimate might be based on desiring only a 2% chance of a false
// alarm every 500 datapoints.
type ThresholdEstimate struct {
	value float64
}

func (te ThresholdEstimate) Changepoint(cost float64) (bool, error) {
	return (cost > te.value), nil
}

func (te ThresholdEstimate) Value() float64 {
	return te.value
}

// NewThresholdEstimate creates a threshold that uses ThresholdEstimate. To
// configure, set the chance of a false changepoint detection to `probability`
// in every N `datapoints`. To tailor this to your data processor, the window
// size and quantile count should be set equal to the same configuration options
// as the processor.
func NewThresholdEstimate(probability float64, datapoints, windowSize, quantiles int) (Threshold, error) {
	if probability <= 0 || probability > 1 {
		return nil, fmt.Errorf("probability must be greater than 0 and less than or equal to 1")
	}
	if windowSize <= 0 {
		return nil, fmt.Errorf("window size must be greater than 0")
	}
	if (datapoints + 1) == windowSize {
		return ThresholdEstimate{value: 1.0 + (2.0 * (math.Sqrt(2.0 * (math.Log((float64(windowSize) * (float64(datapoints) - float64(windowSize) + 1.0)) / probability)))))}, nil
	}
	estimateA := 1.0 - (8.0 * (1.0 / float64(quantiles)) * math.Log(probability/(float64(windowSize)*(float64(datapoints)-float64(windowSize)+1.0))))
	estimateB := 1.0 + (2.0 * (math.Sqrt(2.0 * (math.Log((float64(windowSize) * (float64(datapoints) - float64(windowSize) + 1.0)) / probability)))))
	return ThresholdEstimate{value: math.Max(estimateA, estimateB)}, nil
}

// type AutoNunc struct {
// 	costWindow        *Window[float64]
// 	calibrationCutoff int
// 	thresholdPct      float64
// 	threshold         float64
// }

// func NewAutoNunc(windowSize int, thresholdPct float64) (*AutoNunc, error) {
// 	cw, err := NewWindow[float64](windowSize)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &AutoNunc{
// 		costWindow:        cw,
// 		calibrationCutoff: 0,
// 		thresholdPct:      thresholdPct,
// 		threshold:         -1,
// 	}, nil
// }

// func (an *AutoNunc) Process(processor *NuncProcessor, value float64, quantileCount int) (bool, *Cost, error) {
// 	dataset := an.costWindow.GetFull()
// 	if dataset != nil {
// 		sorted := make([]float64, len(dataset))
// 		copy(sorted, dataset)
// 		sort.Float64s(sorted)
// 		an.threshold = quantile(sorted, an.thresholdPct)
// 		// fmt.Printf("NEW THRESHOLD: %f\n", an.threshold)
// 	}
// 	cost, err := processor.Process(value, quantileCount)
// 	if err == nil {
// 		an.costWindow.Push(cost.Value())
// 	}
// 	if an.threshold >= 0 {
// 		if cost.value > an.threshold {
// 			return true, cost, err
// 		}
// 	}
// 	return false, cost, err
// }
