package nunc

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// NUNC is a container for the NUNC changepoint detection algorithm logic and
// configuration.
type NUNC struct {
	processor *NuncProcessor
	threshold Threshold

	windowSize int
	quantiles  int

	changepoints []uint64
}

// NuncOpt is a generic configuration option for the NUNC algorithm.
type NuncOpt func(n *NUNC) error

// NewNUNC creates a new nunc with the given configuration options. All NUNC
// containers require a window size and quantiles configuration to correctly
// configure the processor. Futher options for configuration such as thresholds
// can be provided too and are executed in order. If the configuration is in any
// way invalid, an error is returned.
func NewNUNC(windowSize, quantiles int, opts ...NuncOpt) (*NUNC, error) {
	processor, err := New(windowSize)
	if err != nil {
		return nil, err
	}
	n := &NUNC{
		processor: processor,
		threshold: nil,

		windowSize: windowSize,
		quantiles:  quantiles,

		changepoints: make([]uint64, 0),
	}
	for idx, opt := range opts {
		if err := opt(n); err != nil {
			return nil, fmt.Errorf("failed to apply opt %d: %s", idx, err.Error())
		}
	}
	return n, nil
}

// Push adds a new datapoint to the NUNC dataset. This is processed using the
// NUNC processor and if a changepoint is detected in the data window, the index
// of the changepoint is returned. If no changepoint is detected, 0 is returned.
// If a changepoint is detected that has already been detected, this is
// suppressed and 0 is returned.
func (n *NUNC) Push(datapoint float64) uint64 {
	if cost, err := n.processor.Process(datapoint, n.quantiles); err == nil && cost != nil {
		if n.threshold != nil {
			if change, err := n.threshold.Changepoint(cost.Value()); err == nil && change {
				if ok := slices.Contains(n.changepoints, cost.Index()); !ok {
					n.changepoints = append(n.changepoints, cost.Index())
					return cost.Index()
				}
			}
		}
	}
	return 0
}

// Threshold returns the current value used as the maximum cost threshold where,
// if exceeded, a datapoint is considered to be a changepoint.
func (n NUNC) Threshold() float64 {
	return n.threshold.Value()
}

// Changepoints returns the datapoint indexes where changepoints are recorded to
// have occurred in the currently processed dataset.
func (n NUNC) Changepoints() []uint64 {
	return n.changepoints
}

// OptThresholdEstimate calculates an appropriate threshold value for the nunc
// configuration. To calculate the estimate, a proability of false alarm per
// 1000 datapoints must be provided.
func OptThresholdEstimate(probability float64) NuncOpt {
	return func(n *NUNC) error {
		threshold, err := NewThresholdEstimate(probability, 1000, n.windowSize, n.quantiles)
		if err != nil {
			return err
		}
		n.threshold = threshold
		return nil
	}
}

// OptThreshold defines a static threshold for the nunc changepoint detection.
func OptThreshold(value float64) NuncOpt {
	return func(n *NUNC) error {
		n.threshold = NewThresholdStatic(value)
		return nil
	}
}
