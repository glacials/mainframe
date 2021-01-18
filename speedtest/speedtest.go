package speedtest

import (
	"fmt"
	"log"

	"github.com/ddo/go-fast"
)

// Run runs a speedtest and records the results.
func Run(logger *log.Logger) error {
	logger = log.New(logger.Writer(), "[speedtest] ", logger.Flags())
	logger.Printf("Starting")

	client := fast.New()
	if err := client.Init(); err != nil {
		return fmt.Errorf("speedtest initialization failed: %v", err)
	}

	urls, err := client.GetUrls()
	if err != nil {
		return fmt.Errorf("speedtest URL fetch failed: %v", err)
	}

	kbpsChan := make(chan float64)
	i := 0
	var kbpsSum float64
	go func() {
		for kbps := range kbpsChan {
			kbpsSum += kbps
			i++
		}
	}()

	if err := client.Measure(urls, kbpsChan); err != nil {
		return fmt.Errorf("speedtest measure failed: %v", err)
	}
	if i == 0 {
		return fmt.Errorf("speedtest didn't get any kbps packets; starting over")
	}
	logger.Printf("Finished: %.2f Mbps", kbpsSum/float64(i)/1000)
	return nil
}
