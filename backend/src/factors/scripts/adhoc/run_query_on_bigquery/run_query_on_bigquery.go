package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"

	C "factors/config"
	BQ "factors/services/bigquery"

	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")
	projectIDFlag := flag.Int64("project_id", 0, "Project ID for the bigquery account")
	queryFlag := flag.String("query", "", "Query to run on Bigquery")
	outputFileFlag := flag.String("outf", "", "If the query result is to be written to a file")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	flag.Parse()

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == 0 {
		log.Fatal("Invalid project ID 0")
	} else if *queryFlag == "" {
		log.Fatal("Query can not be empty")
	}

	log.Info("Starting to initialize database.")
	appName := "push_to_bigquery"
	config := &C.Configuration{
		AppName: appName,
		Env:     *envFlag,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
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
		for _, line := range queryResult {
			fmt.Println(strings.Join(line, " | "))
		}
	}
}
