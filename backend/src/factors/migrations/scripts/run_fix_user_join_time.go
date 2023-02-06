package main

// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_fix_user_join_time.go --project_id=1
// Fix join timestamps of users. Make it lesser than the first event seen for the user.

import (
	C "factors/config"
	//	"factors/model/store/postgres"
	"factors/util"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")
	dryRunFlag := flag.Bool("dry_run", true, "values are updated only when dry_run is false")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#FixUserJoinTime", *env)

	config := &C.Configuration{
		Env: *env,
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	if *projectIdFlag <= 0 {
		log.Fatal("Failed to pull events. Invalid project_id.")
	}
	if err := postgres.GetStore().FixAllUsersJoinTimestampForProject(db, *projectIdFlag, *dryRunFlag); err != nil {
		log.Fatal(err)
	}
}
