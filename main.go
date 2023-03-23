package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "twos.dev/mainframe/coldbrewcrew/iworkout"
	"twos.dev/mainframe/db"
	"twos.dev/mainframe/pottytrainer"
	"twos.dev/mainframe/web"
)

var (
	version     = "development"
	versionFlag = flag.Bool("version", false, "prints mainframe version")
	debugFlag   = flag.Bool(
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

	logger.Printf("Booting mainframe %s", version)
	db, err := db.New(logger, "mainframe")
	if err != nil {
		logger.Fatalf("database error: %v", err)
	}

	mux, err := web.Start(logger, version)
	if err != nil {
		logger.Fatalf("web error: %v", err)
	}

	google, err := newGoogleClient(logger, db, mux)
	if err != nil {
		logger.Fatalf("gcp client error: %v", err)
	}

	pottyMux := http.NewServeMux()
	mux.Handle("/potty/", http.StripPrefix("/potty", pottyMux))
	if err := pottytrainer.Run(logger, pottyMux); err != nil {
		logger.Fatalf("potty trainer error: %v", err)
	}

	if err := startCron(logger, db, version, mux, google); err != nil {
		logger.Fatalf("cron error: %v", err)
	}

	logger.Println("Mainframe booted")
	select {}
}
