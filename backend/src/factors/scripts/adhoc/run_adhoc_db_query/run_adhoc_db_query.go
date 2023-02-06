package main

import (
	C "factors/config"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	flag.Parse()
	appName := "adhoc_db_query"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
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

	appCountQuery := fmt.Sprintf(
		"SELECT JSON_EXTRACT_STRING(users.properties, '$hubspot_deal_amount') FROM users  " +
			"WHERE users.project_id = 399 " +
			"AND join_timestamp>=946688461 " +
			"AND join_timestamp<=1651568362 " +
			"AND source=2 " +
			"AND (is_group_user=0 or is_group_user IS NULL)")
	startTime := U.TimeNowUnix()
	rows, tx, err, _ := store.GetStore().ExecQueryWithContext(appCountQuery, []interface{}{})
	if err != nil {
		log.WithError(err).Error("appCountQuery failed.")
	}
	defer U.CloseReadQuery(rows, tx)
	count := 0
	resultSize := reflect.TypeOf(rows).Size()
	for rows.Next() {
		value := 0.0
		if err = rows.Scan(&value); err == nil {
			count++
			//log.WithFields(log.Fields{"err": err}).Error("appCountQuery Parse failed.")
		}
	}
	endTime := U.TimeNowUnix()
	log.WithFields(
		log.Fields{"result": count, "timeTaken": endTime - startTime,
			"resultSize": resultSize}).Info("appCountQuery completed")

	appSumQuery := fmt.Sprintf(
		"SELECT JSON_EXTRACT_STRING(users.properties, '$hubspot_deal_amount') FROM users  " +
			"WHERE users.project_id = 399 " +
			"AND join_timestamp>=946688461 " +
			"AND join_timestamp<=1651568362 " +
			"AND source=2 " +
			"AND (is_group_user=0 or is_group_user IS NULL)")
	startTime = U.TimeNowUnix()
	rows, tx, err, _ = store.GetStore().ExecQueryWithContext(appSumQuery, []interface{}{})
	if err != nil {
		log.WithError(err).Error("appSumQuery failed.")
	}
	defer U.CloseReadQuery(rows, tx)
	sum := 0.0
	resultSize = reflect.TypeOf(rows).Size()
	for rows.Next() {
		value := 0.0
		if err = rows.Scan(&value); err == nil {
			sum += value
			//log.WithFields(log.Fields{"err": err}).Error("appSumQuery Parse failed.")
		}

	}
	endTime = U.TimeNowUnix()
	log.WithFields(
		log.Fields{"result": sum, "timeTaken": endTime - startTime,
			"resultSize": resultSize}).Info("appSumQuery completed")

	sqlCountQuery := fmt.Sprintf(
		"SELECT COUNT(JSON_EXTRACT_STRING(users.properties, '$hubspot_deal_amount')) FROM users " +
			"WHERE users.project_id = 399 " +
			"AND join_timestamp>=946688461 " +
			"AND join_timestamp<=1651568362 " +
			"AND source=2 " +
			"AND (is_group_user=0 or is_group_user IS NULL)")
	startTime = U.TimeNowUnix()
	rows, tx, err, _ = store.GetStore().ExecQueryWithContext(sqlCountQuery, []interface{}{})
	if err != nil {
		log.WithError(err).Error("sqlCountQuery failed.")
	}
	defer U.CloseReadQuery(rows, tx)
	count = 0
	resultSize = reflect.TypeOf(rows).Size()
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("sqlCountQuery Parse failed.")
		}
	}
	endTime = U.TimeNowUnix()
	log.WithFields(
		log.Fields{"result": count, "timeTaken": endTime - startTime,
			"resultSize": resultSize}).Info("sqlCountQuery completed")

	sqlSumQuery := fmt.Sprintf(
		"SELECT SUM(JSON_EXTRACT_STRING(users.properties, '$hubspot_deal_amount')) FROM users " +
			"WHERE users.project_id = 399 " +
			"AND join_timestamp>=946688461 " +
			"AND join_timestamp<=1651568362 " +
			"AND source=2 " +
			"AND (is_group_user=0 or is_group_user IS NULL)")
	startTime = U.TimeNowUnix()
	rows, tx, err, _ = store.GetStore().ExecQueryWithContext(sqlSumQuery, []interface{}{})
	if err != nil {
		log.WithError(err).Error("sqlSumQuery failed.")
	}
	defer U.CloseReadQuery(rows, tx)
	sum = 0.0
	resultSize = reflect.TypeOf(rows).Size()
	for rows.Next() {
		if err = rows.Scan(&sum); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("sqlSumQuery Parse failed.")
		}
	}
	endTime = U.TimeNowUnix()
	log.WithFields(
		log.Fields{"result": sum, "timeTaken": endTime - startTime,
			"resultSize": resultSize}).Info("sqlSumQuery completed")

}
