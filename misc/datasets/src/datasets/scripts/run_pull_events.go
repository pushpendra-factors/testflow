package main

import (
	C "datasets/config"
	"flag"

	log "github.com/sirupsen/logrus"

	U "datasets/util"
)

func main() {
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	projectId := flag.Uint64("project_id", 0, "")
	startTime := flag.Int64("start_time", 0, "")
	endTime := flag.Int64("end_time", 0, "")
	dir := flag.String("dir", "", "")
	flag.Parse()

	C.InitDB(*dbHost, *dbPort, *dbUser, *dbName, *dbPass)

	if *projectId == 0 {
		log.Fatal("Invalid project_id")
	}

	if *dir == "" {
		log.Fatal("No workspace dir given")
	}

	if *startTime == 0 || *endTime == 0 {
		log.Fatal("Invalid start_time or end_time")
	}

	_, err := U.PullEvents(*projectId, *startTime, *endTime, *dir)
	if err != nil {
		log.WithError(err).Fatal("Failed pulling events for lend.co")
	}
}
