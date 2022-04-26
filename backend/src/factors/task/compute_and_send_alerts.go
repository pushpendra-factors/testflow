package task

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)
type message struct {
	AlertName     string
	AlertType     int
	Operator      string
	ActualValue   float64
	ComparedValue float64
	Value         float64
	DateRange     string
	ComparedTo    string
}

type dateRanges struct {
	from      int64
	to        int64
	prev_from int64
	prev_to   int64
}
func ComputeAndSendAlerts(projectID uint64, configs map[string]interface{}) (map[string]interface{}, bool) {
	allAlerts, errCode := store.GetStore().GetAllAlerts(projectID)
	if errCode != http.StatusFound {
		log.Fatalf("Failed to get all alerts for project_id: %v", projectID)
		return nil, false
	}
	var alertDescription model.AlertDescription
	var alertConfiguration model.AlertConfiguration
	var kpiQuery model.KPIQuery
	var date_range dateRanges

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
			log.Error(err)
			continue
		}
		err = U.DecodePostgresJsonbToStructType(alertDescription.Query, &kpiQuery)
		if err != nil {
			log.Errorf("Error decoding query for project_id: %v, alert_name: %s", projectID, alert.AlertName)
			log.Error(err)
			continue
		}
		timezoneString := U.TimeZoneString(kpiQuery.Timezone)
		date_range, err = getDateRange(timezoneString, alertDescription.DateRange, alertDescription.ComparedTo)
		if err != nil {
			log.Errorf("failed to getDateRange, error: %v for project_id: %v,alert_name: %s,", err, projectID, alert.AlertName)
			continue
		}
		statusCode, actualValue, comparedValue, err := executeAlertsKPIQuery(projectID, alert.AlertType, date_range, kpiQuery)
		if (err != nil) || (statusCode != http.StatusOK) {
			log.Errorf("failed to execute query for project_id: %v, alert_name: %s", projectID, alert.AlertName)
			log.Errorf("status code: %v, error: %s", statusCode, err)
			continue
		}
		value, err := strconv.ParseFloat(alertDescription.Value, 64)
		if err != nil {
			log.Errorf("failed to convert value to float64 for alertName: %s", alert.AlertName)
			continue
		}
		notify, err := sendAlert(alertDescription.Operator, actualValue, comparedValue, value)
		if err != nil {
			log.Errorf("failed to compare results for project_id: %v,alert_name: %s, error: %v", projectID, alert.AlertName, err)
			continue
		}
		if notify {
			msg := message{
				AlertName:     alert.AlertName,
				AlertType:     alert.AlertType,
				Operator:      alertDescription.Operator,
				ActualValue:   actualValue,
				ComparedValue: comparedValue,
				Value:         value,
				DateRange:     alertDescription.DateRange,
				ComparedTo:    alertDescription.ComparedTo,
			}
			if alertConfiguration.IsEmailEnabled {
				sendEmailAlert(msg, alertConfiguration.Emails)
			}
			if alertConfiguration.IsSlackEnabled {
				sendSlackAlert(msg)
			}
		}
	}
	return nil, true
}
func executeAlertsKPIQuery(projectID uint64, alertType int, date_range dateRanges, kpiQuery model.KPIQuery) (statusCode int, actualValue float64, comparedValue float64, err error) {
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	kpiQuery.From = date_range.from
	kpiQuery.To = date_range.to
	kpiQueryGroup.Queries = append(kpiQueryGroup.Queries, kpiQuery)
	results, statusCode := store.GetStore().ExecuteKPIQueryGroup(projectID, "", kpiQueryGroup)
	if len(results) != 1 {
		return statusCode, actualValue, comparedValue, errors.New("empty or invalid result")
	}
	if len(results[0].Rows) == 0 {
		return statusCode, actualValue, comparedValue, errors.New("empty or invalid value in result")
	}
	if statusCode != http.StatusOK {
		return statusCode, actualValue, comparedValue, nil
	}
	actualValue = results[0].Rows[0][0].(float64)
	if alertType == 2 {
		kpiQueryGroup.Queries = []model.KPIQuery{}
		kpiQuery.From = date_range.prev_from
		kpiQuery.To = date_range.prev_to
		kpiQueryGroup.Queries = append(kpiQueryGroup.Queries, kpiQuery)
		results, statusCode = store.GetStore().ExecuteKPIQueryGroup(projectID, "", kpiQueryGroup)
		if len(results) != 1 {
			return statusCode, actualValue, comparedValue, errors.New("empty or invalid result")
		}
		if len(results[0].Rows) == 0 {
			return statusCode, actualValue, comparedValue, errors.New("empty or invalid value in result")
		}
		if statusCode != http.StatusOK {
			return statusCode, actualValue, comparedValue, nil
		}
		comparedValue = results[0].Rows[0][0].(float64)
	}

	return statusCode, actualValue, comparedValue, nil
}

func getDateRange(timezone U.TimeZoneString, dateRange string, prevDateRange string) (dateRanges, error) {

	if dateRange == model.LAST_MONTH || dateRange == model.LAST_QUARTER {
		return dateRanges{}, errors.New("invalid date range")
	}
	var from, to, prev_from, prev_to time.Time
	var err error
	// remove this after cache support
	currentTime := U.TimeNowIn(timezone)
	switch dateRange {
	case model.LAST_WEEK:
		from = currentTime.AddDate(0, 0, -6)
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
		from = currentTime.AddDate(0, 0, 1)
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
		from = currentTime.AddDate(0, 0, 1)
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

func sendAlert(operator string, actualValue float64, comparedValue float64, value float64) (bool, error) {
	switch operator {
	case model.IS_LESS_THAN:
		if actualValue < value {
			return true, nil
		}
	case model.IS_GREATER_THAN:
		if actualValue > value {
			return true, nil
		}
	case model.DECREASED_BY_MORE_THAN:
		if (comparedValue - actualValue) > value {
			return true, nil
		}
	case model.INCREASED_BY_MORE_THAN:
		if (actualValue - comparedValue) > value {
			return true, nil
		}
	case model.INCREASED_OR_DECREASED_BY_MORE_THAN:
		if math.Abs(actualValue-comparedValue) > value {
			return true, nil
		}
	case model.PERCENTAGE_HAS_DECREASED_BY_MORE_THAN:
		if (comparedValue-actualValue)*100 > comparedValue*(value) {
			return true, nil
		}
	case model.PERCENTAGE_HAS_INCREASED_BY_MORE_THAN:
		if (actualValue-comparedValue)*100 > comparedValue*(value) {
			return true, nil
		}
	case model.PERCENTAGE_HAS_INCREASED_OR_DECREASED_BY_MORE_THAN:
		if math.Abs(actualValue-comparedValue)*100 > comparedValue*(value) {
			return true, nil
		}
	default:
		return false, errors.New("invalied comparsion")
	}
	return false, nil
}

func sendEmailAlert(msg message, emails []string) {
	var success, fail int
	sub, text, html := CreateAlertTemplate(msg)
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

func CreateAlertTemplate(msg message) (subject, text, html string) {
	subject = "Alert"
	var comparedValue, comparedDate string
	if msg.AlertType == 2 {
		comparedValue = fmt.Sprintf("compared_value : %v ,", msg.ComparedValue)
		comparedDate = fmt.Sprintf(",compared to : %s ", msg.ComparedTo)
	}
	html = fmt.Sprintf(`<h1>Email Alert</h1><p>This alert is regarding %s <br> actual_value : %v , %s value : %v <br> 
	for date range : %s %s<br></p>`, msg.AlertName, msg.ActualValue, comparedValue, msg.Value, msg.DateRange, comparedDate)
	text = fmt.Sprintf(`<h1>Email Alert</h1><p>This alert is regarding %s <br> actual_value : %v , %s value : %v <br> 
	for date range : %s %s<br></p>`, msg.AlertName, msg.ActualValue, comparedValue, msg.Value, msg.DateRange, comparedDate)
	return subject, text, html
}

func sendSlackAlert(msg message) bool {
	fmt.Println("slack alert sent")
	return true
}