package speedtest

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ddo/go-fast"
)

const insert = `
  INSERT INTO speedtests (
    hostname,
    started_at,
    ended_at,
    kbps_down
  ) VALUES (
    $1, $2, $3, $4
  );
`

// Run runs a speedtest and records the results.
func Run(logger *log.Logger, db *sql.DB) error {
	logger = log.New(logger.Writer(), "[speedtest] ", logger.Flags())

  startedAt := time.Now()

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
	logger.Printf("%.2f Mbps", kbpsSum/float64(i)/1000)

  endedAt := time.Now()

  hostname, err := os.Hostname()
  if err != nil {
    return fmt.Errorf("can't get hostname: %v", err)
  }

  _, err = db.Exec(insert, hostname, startedAt, endedAt, kbpsSum/float64(i))
  if err != nil {
    return fmt.Errorf("speedtest insert failed: %v", err)
  }

	return nil
}
