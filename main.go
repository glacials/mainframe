package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/glacials/mainframe/coldbrewcrew/iworkout"
	"github.com/glacials/mainframe/cron"
	"github.com/glacials/mainframe/web"
)

var (
	version     string = "development"
	versionFlag        = flag.Bool("version", false, "prints mainframe version")
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmsgprefix)

	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}

	logger.Print("Booting mainframe")

	go func() {
		if err := web.Start(logger); err != nil {
			logger.Fatalf("web error: %v", err)
		}
	}()

	go func() {
		if err := cron.Start(logger, version); err != nil {
			logger.Fatalf("cron error: %v", err)
		}
	}()

	logger.Println("Mainframe booted")
	for {
	}
}
