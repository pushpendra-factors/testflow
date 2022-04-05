package main

import (
	"database/sql"
	C "factors/config"
	"factors/model/model"
	"factors/model/store/postgres"
	"factors/util"
	"flag"
	"fmt"
	"strconv"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// go run fill_campaign_id_ad_group_id_in_adwords_documents.go --env=development --db_host=localhost --db_port=5432 --db_user=autometa --db_name=autometa --db_pass=@ut0me7a --start_date=20200101 --end_date=20200102
func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	startDateString := flag.String("start_date", "20200101", "")
	endDateString := flag.String("end_date", "20201101", "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#MigrateAgents", *env)

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
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
	db.LogMode(true)
	defer db.Close()

	projectIDToCustomerAccounts, err := postgres.GetStore().GetAdwordsEnabledProjectIDAndCustomerIDsFromProjectSettings()
	if err != nil {
		log.Fatal("failed in getting projectSettings", err)
	}

	endDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, *endDateString)
	for currentStartDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, *startDateString); endDate.After(currentStartDate); currentStartDate = currentStartDate.AddDate(0, 0, 31*1) {
		var currentEndDate time.Time
		if endDate.After(currentStartDate.AddDate(0, 0, 30*1)) {
			currentEndDate = currentStartDate.AddDate(0, 0, 30*1)
		} else {
			currentEndDate = endDate
		}

		for projectID, customerAccounts := range projectIDToCustomerAccounts {
			for _, customerAccountID := range customerAccounts {
				currentStartDateInt, _ := strconv.ParseInt(currentStartDate.Format(U.DATETIME_FORMAT_YYYYMMDD), 10, 64)
				currentEndDateInt, _ := strconv.ParseInt(currentEndDate.Format(U.DATETIME_FORMAT_YYYYMMDD), 10, 64)

				updateKeywordPerformanceReport(projectID, customerAccountID, currentStartDateInt, currentEndDateInt, db)
				updateAdPerformanceReport(projectID, customerAccountID, currentStartDateInt, currentEndDateInt, db)
				updateAdGroupPerformanceReport(projectID, customerAccountID, currentStartDateInt, currentEndDateInt, db)
				updateCampaignPerformanceReport(projectID, customerAccountID, currentStartDateInt, currentEndDateInt, db)

				updateAdGroupJob(projectID, customerAccountID, currentStartDateInt, currentEndDateInt, db)
				updateCampaignJob(projectID, customerAccountID, currentStartDateInt, currentEndDateInt, db)
				for currentDate := currentStartDate; currentEndDate.After(currentDate); currentDate = currentDate.AddDate(0, 0, 1) {
					currentDateInt, _ := strconv.ParseInt(currentDate.Format(U.DATETIME_FORMAT_YYYYMMDD), 10, 64)
					updateAdPerformanceReportWithCampaignID(projectID, customerAccountID, currentDateInt, db)
				}
			}

		}
	}

}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func updateKeywordPerformanceReport(projectID uint64, customerAccountID string, currentStartDate int64, currentEndDate int64, db *gorm.DB) {
	result := db.Exec("UPDATE adwords_documents SET keyword_id = (value->>'id')::bigint, ad_group_id = (value->>'ad_group_id')::bigint, campaign_id = (value->>'campaign_id')::bigint WHERE project_id = ? AND customer_account_id = ? AND type = ? AND timestamp BETWEEN ? AND ? AND keyword_id IS NULL;", projectID, customerAccountID, model.AdwordsDocumentTypeAlias["keyword_performance_report"], currentStartDate, currentEndDate)

	if result.Error != nil {
		log.Error("Error updating row: " + result.Error.Error())
		log.WithField("customer_acc_id", customerAccountID).WithField(
			"project_id", projectID).Error("There was issue in updating keywordsPerformance")
	}
	time.Sleep(time.Second)
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func updateAdPerformanceReport(projectID uint64, customerAccountID string, currentStartDate int64, currentEndDate int64, db *gorm.DB) {
	result := db.Exec("UPDATE adwords_documents SET ad_id = (value->>'id')::bigint, ad_group_id = (value->>'ad_group_id')::bigint WHERE project_id = ? AND customer_account_id = ? AND type = ? AND timestamp BETWEEN ? AND ? AND ad_id IS NULL;", projectID, customerAccountID, model.AdwordsDocumentTypeAlias["ad_performance_report"], currentStartDate, currentEndDate)

	if result.Error != nil {
		log.Error("Error updating row: " + result.Error.Error())
		log.WithField("customer_acc_id", customerAccountID).WithField(
			"project_id", projectID).Error("There was issue in updating AdPerformance")
	}
	time.Sleep(time.Second)
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func updateAdGroupPerformanceReport(projectID uint64, customerAccountID string, currentStartDate int64, currentEndDate int64, db *gorm.DB) {
	result := db.Exec("UPDATE adwords_documents SET ad_group_id = (value->>'ad_group_id')::bigint, campaign_id = (value->>'campaign_id')::bigint WHERE project_id = ? AND customer_account_id = ? AND type = ? AND timestamp BETWEEN ? AND ? AND ad_group_id IS NULL;", projectID, customerAccountID, model.AdwordsDocumentTypeAlias["ad_group_performance_report"], currentStartDate, currentEndDate)

	if result.Error != nil {
		log.Error("Error updating row: " + result.Error.Error())
		log.WithField("customer_acc_id", customerAccountID).WithField(
			"project_id", projectID).Error("There was issue in updating AdGroupPerformance")
	}
	time.Sleep(time.Second)
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func updateCampaignPerformanceReport(projectID uint64, customerAccountID string, currentStartDate int64, currentEndDate int64, db *gorm.DB) {
	result := db.Exec("UPDATE adwords_documents SET campaign_id = (value->>'campaign_id')::bigint WHERE project_id = ? AND customer_account_id = ? AND type = ? AND timestamp BETWEEN ? AND ? AND campaign_id IS NULL;", projectID, customerAccountID, model.AdwordsDocumentTypeAlias["campaign_performance_report"], currentStartDate, currentEndDate)

	if result.Error != nil {
		log.Error("Error updating row: " + result.Error.Error())
		log.WithField("customer_acc_id", customerAccountID).WithField(
			"project_id", projectID).Error("There was issue in updating CampaignPerformance")
	}
	time.Sleep(time.Second)
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func updateAdGroupJob(projectID uint64, customerAccountID string, currentStartDate int64, currentEndDate int64, db *gorm.DB) {
	result := db.Exec("UPDATE adwords_documents SET ad_group_id = (value->>'id')::bigint, campaign_id = (value->>'campaign_id')::bigint WHERE project_id = ? AND customer_account_id = ? AND type = ? AND timestamp BETWEEN ? AND ? AND ad_group_id IS NULL;", projectID, customerAccountID, model.AdwordsDocumentTypeAlias["ad_groups"], currentStartDate, currentEndDate)

	if result.Error != nil {
		log.Error("Error updating row: " + result.Error.Error())
		log.WithField("customer_acc_id", customerAccountID).WithField(
			"project_id", projectID).Error("There was issue in updating AdGroupJob")
	}
	time.Sleep(time.Second)
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func updateCampaignJob(projectID uint64, customerAccountID string, currentStartDate int64, currentEndDate int64, db *gorm.DB) {
	result := db.Exec("UPDATE adwords_documents SET campaign_id = (value->>'id')::bigint WHERE project_id = ? AND customer_account_id = ? AND type = ? AND timestamp BETWEEN ? AND ? AND campaign_id IS NULL;", projectID, customerAccountID, model.AdwordsDocumentTypeAlias["campaigns"], currentStartDate, currentEndDate)

	if result.Error != nil {
		log.Error("Error updating row: " + result.Error.Error())
		log.WithField("customer_acc_id", customerAccountID).WithField(
			"project_id", projectID).Error("There was issue in updating CampaignJob")
	}
	time.Sleep(time.Second)
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func updateAdPerformanceReportWithCampaignID(projectID uint64, customerAccountID string, currentDate int64, db *gorm.DB) {
	rows, err := db.Raw("SELECT campaign_id, id FROM adwords_documents WHERE project_id = ? AND customer_account_id = ? AND type = ? AND timestamp = ? AND id IS NULL;", projectID, customerAccountID, model.AdwordsDocumentTypeAlias["ad_groups"], currentDate).Rows()
	if err != nil {
		log.WithField("customer_acc_id", customerAccountID).WithField(
			"project_id", projectID).Error("There was an issue in fetching campaignid, adgroupid")
		return
	}
	defer rows.Close()

	mapOfCampaignIDAndAdGroupID := getMapOfCampaignIDAndAdGroupID(rows)

	for campaignID, adGroupIDs := range mapOfCampaignIDAndAdGroupID {
		result := db.Exec("UPDATE adwords_documents SET campaign_id = ? WHERE project_id = ? AND customer_account_id = ? AND type = ? AND timestamp = ? AND ad_group_id IN (?) AND campaign_id IS NULL;", campaignID, projectID, customerAccountID, model.AdwordsDocumentTypeAlias["ad_performance_report"], currentDate, adGroupIDs)

		if result.Error != nil {
			log.Error("Error updating row: " + result.Error.Error())
			log.WithField("customer_acc_id", customerAccountID).WithField(
				"project_id", projectID).Error("There was issue in updating campaignId in AdPerformance")
		}
	}
}

func getMapOfCampaignIDAndAdGroupID(rows *sql.Rows) map[uint64][]uint64 {
	mapOfCampaignIDAndAdGroupID := make(map[uint64][]uint64)
	var campaignID, adGroupID uint64
	for rows.Next() {
		if err := rows.Scan(&campaignID, &adGroupID); err != nil {
			log.WithError(err).Error("Failed to get campaignIds, AdgroupIds. Scanning failed.")
			continue
		}
		val, isPresent := mapOfCampaignIDAndAdGroupID[campaignID]
		if isPresent {
			val = append(val, adGroupID)
		} else {
			val = []uint64{adGroupID}
		}
		mapOfCampaignIDAndAdGroupID[campaignID] = val
	}
	return mapOfCampaignIDAndAdGroupID

}
