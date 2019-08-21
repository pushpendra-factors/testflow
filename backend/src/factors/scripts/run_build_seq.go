package main

import (
	C "factors/config"
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	"flag"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", "development", "")
	etcd := flag.String("etcd", "localhost:2379",
		"Comma separated list of etcd endpoints localhost:2379,localhost:2378")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")
	numRoutinesFlag := flag.Int("num_routines", 3, "No of routines")
	projectIdFlag := flag.Uint64("project_id", 0, "Optional: Project Id.")
	projectIdsToSkipFlag := flag.String("project_ids_to_skip", "", "Optional: Comma separated values of projects to skip")
	maxModelSizeFlag := flag.Int64("max_size", 20000000000, "Max size of the model")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	flag.Parse()

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	// init DB, etcd
	config := &C.Configuration{
		Env:           *envFlag,
		EtcdEndpoints: strings.Split(*etcd, ","),
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}

	C.InitConf(config.Env)

	// db is used by M.GetEventNames to build eventInfo.
	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}
	db := C.GetServices().Db
	defer db.Close()

	err = C.InitEtcd(config.EtcdEndpoints)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize etcd")
	}
	etcdClient, err := serviceEtcd.New([]string{*etcd})
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to init etcd client")
	}

	log.WithFields(log.Fields{
		"Env":             *envFlag,
		"EtcdEndpoints":   *etcd,
		"localDiskTmpDir": *localDiskTmpDirFlag,
		"Bucket":          *bucketName,
		"NumRoutines":     *numRoutinesFlag,
	}).Infoln("Initialising")

	if *numRoutinesFlag < 1 {
		log.Fatal("num_routines is less than one.")
	}

	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketName)
	} else {
		cloudManager, err = serviceGCS.New(*bucketName)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
	}
	projectIdToSkipStrings := strings.Split(*projectIdsToSkipFlag, ",")
	projectIdsToSkip := make(map[uint64]bool)
	for _, pid := range projectIdToSkipStrings {
		if pid == "" {
			continue
		}
		if pidUint, err := strconv.ParseUint(pid, 10, 64); err == nil {
			projectIdsToSkip[pidUint] = true
		} else {
			log.WithError(err).Errorln("Failed to parse projects to skip")
			panic(err)
		}
	}

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	T.BuildSequential(*envFlag, db, &cloudManager, etcdClient, diskManager,
		*bucketName, *numRoutinesFlag, *projectIdFlag, projectIdsToSkip, *maxModelSizeFlag)
}
