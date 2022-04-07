package main

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

type message struct {
	Notify        bool
	Type          int
	Operator      string
	ActualValue   int64
	ComparedValue int64
	Value         int64
}

type dateRanges struct {
	from      int64
	to        int64
	prev_from int64
	prev_to   int64
}

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
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")
	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	projectIdFlag := flag.String("project_id", "", "Comma separated list of project ids to run")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}
	defer U.NotifyOnPanic("Script#run_alerts", *env)
	appName := "run_alerts"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
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
		AWSKey:           *awsAccessKeyId,
		AWSSecret:        *awsSecretAccessKey,
		AWSRegion:        *awsRegion,
		EmailSender:      *factorsEmailSender,
	}
	C.InitConf(config)
	C.InitSenderEmail(C.GetFactorsSenderEmail())
	C.InitMailClient(config.AWSKey, config.AWSSecret, config.AWSRegion)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()
	//Initialized configs

	query := "select DISTINCT(project_id) from alerts;"
	rows, err := db.Raw(query).Rows()
	if err != nil {
		log.Fatal(err)
	}
	allDistinctProjectid := make([]uint64, 0)
	for rows.Next() {
		var projectId uint64
		rows.Scan(&projectId)
		allDistinctProjectid = append(allDistinctProjectid, projectId)
	}
	runAllProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
	projectIdsArray := make([]uint64, 0)
	if runAllProjects || *projectIdFlag == "" {
		projectIdsArray = append(projectIdsArray, allDistinctProjectid...)
	} else {
		for projectId := range projectIdsToRun {
			projectIdsArray = append(projectIdsArray, projectId)
		}
	}
	for _, projectID := range projectIdsArray {
		allAlerts, errCode := store.GetStore().GetAllAlerts(projectID)
		if errCode != http.StatusFound {
			log.Fatalf("Failed to get all alerts for project_id: %v", projectID)
			return
		}
		var alertDescription model.AlertDescription
		var alertConfiguration model.AlertConfiguration
		var kpiQuery model.KPIQuery

		for _, alert := range allAlerts {

			err := U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
			if err != nil {
				log.Errorf("failed to decode alert description for project_id: %v, alert_name: %s", projectID, alert.AlertName)
				log.Error(err)
				continue
			}
			err = U.DecodePostgresJsonbToStructType(alert.AlertConfiguration, &alertConfiguration)
			if err != nil {
				log.Errorf("failed to decode alert configuration for project_id: %v, alert_name: %s", projectID, alert.AlertName)
				continue
			}
			err = U.DecodePostgresJsonbToStructType(alertDescription.Query, &kpiQuery)
			if err != nil {
				log.Errorf("Error decoding query for project_id: %v, alert_name: %s", projectID, alert.AlertName)
				continue
			}
			kpiQueryGroup := model.KPIQueryGroup{
				Class:         "kpi",
				Queries:       []model.KPIQuery{},
				GlobalFilters: []model.KPIFilter{},
				GlobalGroupBy: []model.KPIGroupBy{},
			}
			if alert.AlertType == 1 {
				date_range, err := getDateRange(alertDescription.DateRange, "")
				if err != nil {
					log.Errorf("error: %v for project_id: %v,alert_name: %s,", err, projectID, alert.AlertName)
					continue
				}
				kpiQuery.From = date_range.from
				kpiQuery.To = date_range.to
				kpiQueryGroup.Queries = append(kpiQueryGroup.Queries, kpiQuery)
			} else if alert.AlertType == 2 {
				date_range, err := getDateRange(alertDescription.DateRange, alertDescription.ComparedTo)
				if err != nil {
					log.Errorf("error: %v for project_id: %v,alert_name: %s,", err, projectID, alert.AlertName)
					continue
				}
				kpiQuery.From = date_range.from
				kpiQuery.To = date_range.to
				kpiQueryGroup.Queries = append(kpiQueryGroup.Queries, kpiQuery)
				kpiQuery.From = date_range.prev_from
				kpiQuery.To = date_range.prev_to
				kpiQueryGroup.Queries = append(kpiQueryGroup.Queries, kpiQuery)
			}
			results, statusCode := store.GetStore().ExecuteKPIQueryGroup(projectID, "", kpiQueryGroup)
			if statusCode != http.StatusOK {
				log.Errorf("failed to execute query for project_id: %v, alert_name: %s", projectID, alert.AlertName)
				continue
			}
			value, err := strconv.ParseInt(alertDescription.Value, 10, 64)
			if err != nil {
				log.Errorf("failed to convert value to int64 for alertName: %s", alert.AlertName)
				continue
			}
			msg, err := compareResults(alertDescription.Operator, results, value)
			if err != nil {
				log.Errorf("failed to compare results for project_id: %v,alert_name: %s, error: %v", projectID, alert.AlertName, err)
				continue
			}
			if msg.Notify {
				if alertConfiguration.IsEmailEnabled {
					sendEmailAlert(alert.AlertName, msg, alertConfiguration.Emails)
				}
				if alertConfiguration.IsSlackEnabled {
					sendSlackAlert(msg)
				}
			}
		}
	}
}

func getDateRange(dateRange string, prevDateRange string) (dateRanges, error) {
	var from, to, prev_from, prev_to time.Time
	var err error
	// remove this after cache support
	if dateRange == model.LAST_MONTH || dateRange == model.LAST_QUARTER{
		return dateRanges{}, errors.New("invalid date range")
	} 
	switch dateRange {
	case model.LAST_WEEK:
		from = time.Now().AddDate(0, 0, -6)
		from = now.New(from).BeginningOfWeek()
		to = now.New(from).EndOfWeek()
		if prevDateRange == model.PREVIOUS_PERIOD {
			prev_from = from.AddDate(0, 0, -7)
			prev_from = now.New(prev_from).BeginningOfWeek()
			prev_to = now.New(prev_from).EndOfWeek()
		} else if prevDateRange == model.SAME_PERIOD_LAST_YEAR {
			prev_from = from.AddDate(-1, 0, 0)
			prev_from = now.New(prev_from).BeginningOfWeek()
			prev_to = now.New(prev_from).EndOfWeek()
		}
	case model.LAST_MONTH:
		from = time.Now().AddDate(0, 0, 1)
		from = now.New(from).BeginningOfMonth().AddDate(0, 0, -1)
		from = now.New(from).BeginningOfMonth()
		to = now.New(from).EndOfMonth()
		if prevDateRange == model.PREVIOUS_PERIOD {
			prev_from = from.AddDate(0, 0, -1)
			prev_from = now.New(prev_from).BeginningOfMonth()
			prev_to = now.New(prev_from).EndOfMonth()
		} else if prevDateRange == model.SAME_PERIOD_LAST_YEAR {
			prev_from = from.AddDate(-1, 0, 0)
			prev_from = now.New(prev_from).BeginningOfMonth()
			prev_to = now.New(prev_from).EndOfMonth()
		}
	case model.LAST_QUARTER:
		from = time.Now().AddDate(0, 0, 1)
		from = now.New(from).BeginningOfMonth().AddDate(0, 0, -1)
		from = now.New(from).BeginningOfQuarter()
		to = now.New(from).EndOfQuarter()
		if prevDateRange == model.PREVIOUS_PERIOD {
			prev_from = from.AddDate(0, 0, -1)
			prev_from = now.New(prev_from).BeginningOfQuarter()
			prev_to = now.New(prev_from).EndOfQuarter()
		} else if prevDateRange == model.SAME_PERIOD_LAST_YEAR {
			prev_from = from.AddDate(-1, 0, 0)
			prev_from = now.New(prev_from).BeginningOfQuarter()
			prev_to = now.New(prev_from).EndOfQuarter()
		}
	default:
		err = errors.New("invalid date_range")
	}
	return dateRanges{
		from:      from.Unix(),
		to:        to.Unix(),
		prev_from: prev_from.Unix(),
		prev_to:   prev_to.Unix(),
	}, err
}

func abs(val int64) int64 {
	if val < 0 {
		return -val
	}
	return val
}
func compareResults(operator string, results []model.QueryResult, value int64) (message, error) {

	if len(results) == 0 {
		return message{Notify: false}, errors.New("empty result")
	}
	switch operator {
	case model.IS_LESS_THAN:
		if len(results) != 1 {
			return message{Notify: false}, errors.New("invalid query result")
		}
		if len(results[0].Rows) == 0 {
			return message{Notify: false}, errors.New("empty query result")
		}
		actualValue := int64(results[0].Rows[0][0].(float64))
		if actualValue < value {
			return message{
				Notify:      true,
				Type:        1,
				ActualValue: actualValue,
				Value:       value,
				Operator:    operator,
			}, nil
		}
	case model.IS_GREATER_THAN:
		if len(results) != 1 {
			return message{Notify: false}, errors.New("invalid query result")
		}
		if len(results[0].Rows) == 0 {
			return message{Notify: false}, errors.New("empty query result")
		}
		actualValue := int64(results[0].Rows[0][0].(float64))
		if actualValue > value {
			return message{
				Notify:      true,
				Type:        1,
				ActualValue: actualValue,
				Value:       value,
				Operator:    operator,
			}, nil
		}
	case model.DECREASED_BY_MORE_THAN:
		if len(results) != 2 {
			return message{Notify: false}, errors.New("invalid query result")
		}
		if len(results[0].Rows) == 0 || len(results[1].Rows) == 0 {
			return message{Notify: false}, errors.New("empty query result")
		}
		actualValue := int64(results[0].Rows[0][0].(float64))
		comparedValue := int64(results[1].Rows[0][0].(float64))
		if (comparedValue - actualValue) > value {
			return message{
				Notify:        true,
				Type:          2,
				ActualValue:   actualValue,
				ComparedValue: comparedValue,
				Value:         value,
				Operator:      operator,
			}, nil
		}
	case model.INCREASED_BY_MORE_THAN:
		if len(results) != 2 {
			return message{Notify: false}, errors.New("invalid query result")
		}
		if len(results[0].Rows) == 0 || len(results[1].Rows) == 0 {
			return message{Notify: false}, errors.New("empty query result")
		}
		actualValue := int64(results[0].Rows[0][0].(float64))
		comparedValue := int64(results[1].Rows[0][0].(float64))
		if (actualValue - comparedValue) > value {
			return message{
				Notify:        true,
				Type:          2,
				ActualValue:   actualValue,
				ComparedValue: comparedValue,
				Value:         value,
				Operator:      operator,
			}, nil
		}
	case model.INCREASED_OR_DECREASED_BY_MORE_THAN:
		if len(results) != 2 {
			return message{Notify: false}, errors.New("invalid query result")
		}
		if len(results[0].Rows) == 0 || len(results[1].Rows) == 0 {
			return message{Notify: false}, errors.New("empty query result")
		}
		actualValue := int64(results[0].Rows[0][0].(float64))
		comparedValue := int64(results[1].Rows[0][0].(float64))
		if abs(actualValue-comparedValue) > value {
			return message{
				Notify:        true,
				Type:          2,
				ActualValue:   actualValue,
				ComparedValue: comparedValue,
				Value:         value,
				Operator:      operator,
			}, nil
		}
	case model.PERCENTAGE_HAS_DECREASED_BY_MORE_THAN:
		if len(results) != 2 {
			return message{Notify: false}, errors.New("invalid query result")
		}
		if len(results[0].Rows) == 0 || len(results[1].Rows) == 0 {
			return message{Notify: false}, errors.New("empty query result")
		}
		actualValue := int64(results[0].Rows[0][0].(float64))
		comparedValue := int64(results[1].Rows[0][0].(float64))
		if (comparedValue-actualValue)*100 > comparedValue*(value) {
			return message{
				Notify:        true,
				Type:          2,
				ActualValue:   actualValue,
				ComparedValue: comparedValue,
				Value:         value,
				Operator:      operator,
			}, nil
		}
	case model.PERCENTAGE_HAS_INCREASED_BY_MORE_THAN:
		if len(results) != 2 {
			return message{Notify: false}, errors.New("invalid query result")
		}
		if len(results[0].Rows) == 0 || len(results[1].Rows) == 0 {
			return message{Notify: false}, errors.New("empty query result")
		}
		actualValue := int64(results[0].Rows[0][0].(float64))
		comparedValue := int64(results[1].Rows[0][0].(float64))
		if (actualValue-comparedValue)*100 > comparedValue*(value) {
			return message{
				Notify:        true,
				Type:          2,
				ActualValue:   actualValue,
				ComparedValue: comparedValue,
				Value:         value,
				Operator:      operator,
			}, nil
		}
	case model.PERCENTAGE_HAS_INCREASED_OR_DECREASED_BY_MORE_THAN:
		if len(results) != 2 {
			return message{Notify: false}, errors.New("invalid query result")
		}
		if len(results[0].Rows) == 0 || len(results[1].Rows) == 0 {
			return message{Notify: false}, errors.New("empty query result")
		}
		actualValue := int64(results[0].Rows[0][0].(float64))
		comparedValue := int64(results[1].Rows[0][0].(float64))
		if abs(actualValue-comparedValue)*100 > comparedValue*(value) {
			return message{
				Notify:        true,
				Type:          2,
				ActualValue:   actualValue,
				ComparedValue: comparedValue,
				Value:         value,
				Operator:      operator,
			}, nil
		}
	default:
		return message{Notify: false}, errors.New("invalied comparsion")
	}
	return message{Notify: false}, errors.New("invalied comparsion")
}

func sendEmailAlert(alertName string, msg message, emails []string) {
	var success, fail int
	sub, text, html := CreateAlertTemplate(alertName, msg)
	for _, email := range emails {
		err := C.GetServices().Mailer.SendMail(email, C.GetFactorsSenderEmail(), sub, html, text)
		if err != nil {
			fail++
			log.WithError(err).Error("failed to send email alert")
			return
		}
		success++
	}
	log.Info("sent email alert to ", success, "failed to send email alert to ", fail)
}

func CreateAlertTemplate(alertName string, msg message) (subject, text, html string) {
	subject = "Alert"
	html = fmt.Sprintf(`<h1>Email Alert</h1><p>This alert is regarding %s <br> </p>
`, alertName)
	text = fmt.Sprintf(`<p>This alert is regarding %s <br> </p>
`, alertName)
	return subject, text, html
}

func sendSlackAlert(msg message) bool {
	fmt.Println("slack alert sent")
	return true
}
