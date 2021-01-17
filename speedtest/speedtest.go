package speedtest

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ddo/go-fast"
)

var logger = log.New(os.Stderr, "[speedtest]", log.Ldate|log.Ltime)

const interval = 24 * time.Hour

// Monitor only terminates if it hits an error, so should be run in a goroutine.
func Monitor(logger *log.Logger) error {
	logger = log.New(logger.Writer(), "[speedtest] ", logger.Flags())
	logger.Printf("Starting speedtest daemon")

	client := fast.New()
	if err := client.Init(); err != nil {
		return fmt.Errorf("speedtest initialization failed: %v", err)
	}

	for {
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
			logger.Fatalf("speedtest didn't get any kbps packets; starting over")
			continue
		}
		logger.Printf("Bandwidth: %d Mbps", int(kbpsSum)/i/1000)
		time.Sleep(interval)
	}
}
