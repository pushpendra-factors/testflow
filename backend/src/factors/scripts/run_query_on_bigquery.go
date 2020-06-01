package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	C "factors/config"
	BQ "factors/services/bigquery"

	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", "development", "Environment. Could be development|staging|production.")
	projectIDFlag := flag.Uint64("project_id", 0, "Project ID for the bigquery account")
	queryFlag := flag.String("query", "", "Query to run on Bigquery")
	outputFileFlag := flag.String("outf", "", "If the query result is to be written to a file")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	flag.Parse()

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == 0 {
		log.Fatal("Invalid project ID 0")
	} else if *queryFlag == "" {
		log.Fatal("Query can not be empty")
	}

	log.Info("Starting to initialize database.")
	config := &C.Configuration{
		AppName: "script_push_to_bigquery",
		Env:     *envFlag,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}
	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	ctx := context.Background()
	client, err := BQ.CreateBigqueryClientForProject(&ctx, *projectIDFlag)
	if err != nil {
		log.WithError(err).Error("Failed to get bigquery client")
	}
	defer client.Close()

	var queryResult [][]string
	err = BQ.ExecuteQuery(&ctx, client, *queryFlag, &queryResult)
	if err != nil {
		log.WithError(err).Error("Error while executing query")
	}
	if *outputFileFlag != "" {
		outf, err := os.Create(*outputFileFlag)
		if err != nil {
			log.WithError(err).Error("Failed to open file")
			fmt.Println(queryResult)
			return
		}
		defer outf.Close()

		csvWriter := csv.NewWriter(outf)
		defer csvWriter.Flush()

		csvWriter.WriteAll(queryResult)
	} else {
		fmt.Println(queryResult)
	}
}
