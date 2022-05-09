package task

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	AlertName     string
	AlertType     int
	Operator      string
	Category      string
	ActualValue   float64
	ComparedValue float64
	Value         float64
	DateRange     string
	ComparedTo    string
	From          int64
	To            int64
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
	var dateRange dateRanges
	status := make(map[string]interface{})
	endTimestampUnix := configs["endTimestamp"].(int64)
	if endTimestampUnix == 0 || endTimestampUnix > U.TimeNowUnix() {
		status["error"] = "invalid end timestamp"
		return status, false
	}
	endTimestamp := time.Unix(endTimestampUnix, 0)
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
		dateRange, err = getDateRange(timezoneString, alertDescription.DateRange, alertDescription.ComparedTo, endTimestamp)
		if err != nil {
			log.Errorf("failed to getDateRange, error: %v for project_id: %v,alert_name: %s,", err, projectID, alert.AlertName)
			continue
		}
		statusCode, actualValue, comparedValue, err := executeAlertsKPIQuery(projectID, alert.AlertType, dateRange, kpiQuery)
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
			msg := Message{
				AlertName:     alertDescription.Name,
				AlertType:     alert.AlertType,
				Operator:      alertDescription.Operator,
				Category:      kpiQuery.DisplayCategory,
				ActualValue:   actualValue,
				ComparedValue: comparedValue,
				Value:         value,
				DateRange:     alertDescription.DateRange,
				ComparedTo:    alertDescription.ComparedTo,
				From:          dateRange.from,
				To:            dateRange.to,
			}
			if alertConfiguration.IsEmailEnabled {
				sendEmailAlert(msg, dateRange, timezoneString, alertConfiguration.Emails)
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

func getDateRange(timezone U.TimeZoneString, dateRange string, prevDateRange string, endTimeStamp time.Time) (dateRanges, error) {

	if dateRange == model.LAST_MONTH || dateRange == model.LAST_QUARTER {
		return dateRanges{}, errors.New("invalid date range")
	}
	var from, to, prev_from, prev_to time.Time
	var err error
	endTimeStamp = U.ConvertTimeIn(endTimeStamp, timezone)
	switch dateRange {
	case model.LAST_WEEK:
		// end time stamp from config
		from = endTimeStamp.AddDate(0, 0, -6)
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
		from = endTimeStamp.AddDate(0, 0, 1)
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
		from = endTimeStamp.AddDate(0, 0, 1)
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

func sendEmailAlert(msg Message, dateRange dateRanges, timezone U.TimeZoneString, emails []string) {
	var success, fail int
	sub := "Factors Alert"
	text := ""
	var statement string
	fromTime := time.Unix(dateRange.from, 0)
	toTime := time.Unix(dateRange.to, 0)

	fromTime = U.ConvertTimeIn(fromTime, timezone)
	toTime = U.ConvertTimeIn(toTime, timezone)

	year, month, day := fromTime.Date()
	from := fmt.Sprintf("%d-%d-%d", day, month, year)

	year, month, day = toTime.Date()
	to := fmt.Sprintf("%d-%d-%d", day, month, year)

	if msg.AlertType == 1 {
		statement = fmt.Sprintf(`%s %s recorded for %s in %s from %s to %s`, fmt.Sprint(msg.ActualValue), strings.ReplaceAll(msg.AlertName, "_", " "), strings.ReplaceAll(msg.Category, "_", " "), strings.ReplaceAll(msg.DateRange, "_", " "), from, to)
	} else if msg.AlertType == 2 {
		statement = fmt.Sprintf(`%s %s %s for %s in %s (from %s to %s ) compared to %s - %s(%s)`, strings.ReplaceAll(msg.AlertName, "_", " "), strings.ReplaceAll(msg.Operator, "_", " "), fmt.Sprint(msg.Value), strings.ReplaceAll(msg.Category, "_", " "), strings.ReplaceAll(msg.DateRange, "_", " "), from, to, strings.ReplaceAll(msg.ComparedTo, "_", " "), fmt.Sprint(msg.ActualValue), fmt.Sprint(msg.ComparedValue))
	}
	html := U.CreateAlertTemplate(statement)
	dryRunFlag := C.GetConfig().EnableDryRunAlerts
	if dryRunFlag {
		log.Info("Dry run mode enabled. No emails will be sent")
		log.Info(html)
		return
	}
	for _, email := range emails {
		err := C.GetServices().Mailer.SendMail(email, C.GetFactorsSenderEmail(), sub, html, text)
		if err != nil {
			fail++
			log.WithError(err).Error("failed to send email alert")
			return
		}
		success++
	}
	log.Info("sent email alert to ", success, " failed to send email alert to ", fail)
}

func sendSlackAlert(msg Message) bool {
	fmt.Println("slack alert sent")
	return true
}
