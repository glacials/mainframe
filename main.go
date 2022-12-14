package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "twos.dev/mainframe/coldbrewcrew/iworkout"
	"twos.dev/mainframe/cron"
	"twos.dev/mainframe/db"
	"twos.dev/mainframe/web"
)

var (
	version     string = "development"
	versionFlag        = flag.Bool("version", false, "prints mainframe version")
	debugFlag          = flag.Bool(
		"debug",
		false,
		"runs in debug mode (frequent crons)",
	)
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmsgprefix)

	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}

	logger.Print("Booting mainframe")
	db, err := db.New(logger, "mainframe")
	if err != nil {
		logger.Fatalf("database error: %v", err)
	}

	go func() {
		if err := web.Start(logger); err != nil {
			logger.Fatalf("web error: %v", err)
		}
	}()

	if err := cron.Start(logger, db, version); err != nil {
		logger.Fatalf("cron error: %v", err)
	}

	logger.Println("Mainframe booted")
	select {}
}
