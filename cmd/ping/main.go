package main

import (
	"fmt"
	"time"

	"github.com/go-ping/ping"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/willfantom/go-nunc"
)

const (
	toolName string = "nunc-ping"
)

var (
	printChan chan string = make(chan string)

	count    int = 0
	interval int = 1000

	verbose bool = false

	windowSize       int     = 300
	quantiles        int     = 3
	falseProbability float64 = 0.02
)

var (
	rootCmd = &cobra.Command{
		Use:   toolName,
		Short: "Check for changes in round-trip latency distribution",
		Long: `Use the NUNC changepoint detection algorithm to detect changes in round-trip
latency distribution between 2 contactable points via a modified ping command`,
		Args: cobra.ExactArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetLevel(logrus.InfoLevel)
			if verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			pinger, err := ping.NewPinger(args[0])
			if err != nil {
				logrus.WithError(err).WithField("targetAddress", args[0]).Fatalln("failed to create pinger with the target address")
			}
			logrus.WithField("targetAddress", args[0]).Debugln("created pinger with a target address")
			if count != 0 {
				pinger.Count = count
			}
			if interval != 0 {
				pinger.Interval = time.Duration(interval * int(time.Millisecond))
			}

			nunc, err := nunc.NewNUNC(windowSize, quantiles, nunc.OptThresholdEstimate(falseProbability))
			if err != nil {
				panic(err)
			}

			pinger.OnRecv = func(s *ping.Packet) {
				cp := nunc.Push(s.Rtt.Seconds())
				if cp != 0 {
					printChan <- fmt.Sprintf("Changepoint at %d || Average RTT: %s", cp, s.Rtt)
				}
			}
			go printLoop()
			pinger.Run()
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func printLoop() {
	for {
		message := <-printChan
		fmt.Println(message)
	}
}

func init() {

	rootCmd.PersistentFlags().IntVarP(&count, "count", "c", count, "number of ping to perform before exiting or 0 for infinite")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", interval, "milliseconds between each ping")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", verbose, "print debug logs")

	rootCmd.PersistentFlags().Float64VarP(&falseProbability, "false-proability", "p", falseProbability, "chance of a false changepoint being detected per 1000 datapoints")
	rootCmd.PersistentFlags().IntVarP(&windowSize, "window", "w", windowSize, "size of the nunc detection window")
	rootCmd.PersistentFlags().IntVarP(&quantiles, "quantiles", "q", quantiles, "number of quantiles to use when running nunc")

}
