package main

import (
	C "factors/config"
	M "factors/model/model"
	U "factors/util"
	"flag"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	appName := "insert_id_text"
	flag.Parse()
	config := &C.Configuration{
		Env: *env,
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
		log.Fatal("Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	var count int = 0
	var Queries []M.Queries
	err = db.Select("id, id_text").Find(&Queries).Error
	if err != nil {
		log.Error(err)
		return
	}
	for _, Query := range Queries {
		var id int64 = Query.ID
		if Query.IdText == "" {
			id_text := U.RandomStringForSharableQuery(40)
			var queries M.Queries
			db.Model(&queries).Where("id =?", id).Update("id_text", id_text)
			if err != nil {
				log.Error(err)
				continue
			}
			count++
		}

	}
	log.Info("succesfully added text_id to " + strconv.Itoa(count) + " records")
}
