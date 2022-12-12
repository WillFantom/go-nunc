package main

import (
	"fmt"
	"time"

	"github.com/willfantom/go-nunc"
)

const (
	windowSize       int     = 250
	quantileCount    int     = 3
	falseProbability float64 = 0.02
)

func main() {

	// Create dataset (changepoints 700, 1300, 2000, 2400)
	dist := append(NormalDist(1, 0, 700), NormalDist(0.5, 20, 600)...)
	dist = append(dist, NormalDist(4, 5, 700)...)
	dist = append(dist, NormalDist(95, 6000, 400)...)
	dist = append(dist, NormalDist(95, 6000, 5000)...)

	// Create NUNC with an estimated threshold
	n, err := nunc.NewNUNC(windowSize, quantileCount, nunc.OptThresholdEstimate(falseProbability))
	if err != nil {
		panic(err)
	}

	// Output the estimated cost threshold
	fmt.Printf("Estimated cost threshold: %f\n", n.Threshold())

	// Start timer
	start := time.Now()

	// Run NUNC on the dataset
	for _, datapoint := range dist {
		if changepoint := n.Push(datapoint); changepoint > 0 {
			fmt.Printf("Changepoint detected at index %d\n", changepoint)
		}
	}

	// Output total processing time
	fmt.Printf("Executed NUNC on %d datapoints in %s\n", len(dist), time.Since(start))
}
