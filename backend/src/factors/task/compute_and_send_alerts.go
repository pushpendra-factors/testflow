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
	"strings"
	"time"

	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	AlertName     string
	AlertType     int
	Operator      string
	Category      string
	ActualValue   float64
	ComparedValue float64
	PageURL       string
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
	var dateRange dateRanges
	status := make(map[string]interface{})
	endTimestampUnix := configs["endTimestamp"].(int64)
	if endTimestampUnix == 0 || endTimestampUnix > U.TimeNowUnix() {
		status["error"] = "invalid end timestamp"
		return status, false
	}
	endTimestamp := time.Unix(endTimestampUnix, 0)
	for _, alert := range allAlerts {
		var kpiQuery model.KPIQuery
		alert.LastRunTime = time.Now()

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
		if kpiQuery.Category == model.ProfileQueryClass {
			kpiQuery.GroupByTimestamp = getGBTForKPIQuery(alertDescription.DateRange)
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
				AlertName:     strings.Title(alertDescription.Name),
				AlertType:     alert.AlertType,
				Operator:      alertDescription.Operator,
				Category:      strings.Title(filterStringbyLastWord(kpiQuery.DisplayCategory, "metrics")),
				ActualValue:   actualValue,
				ComparedValue: comparedValue,
				PageURL:       kpiQuery.PageUrl,
				Value:         value,
				DateRange:     alertDescription.DateRange,
				ComparedTo:    alertDescription.ComparedTo,
				From:          dateRange.from,
				To:            dateRange.to,
			}
			if alertConfiguration.IsEmailEnabled {
				sendEmailAlert(projectID, msg, dateRange, timezoneString, alertConfiguration.Emails)
			}
			if alertConfiguration.IsSlackEnabled {
				sendSlackAlert(msg)
			}
		}
		alert.LastAlertSent = true
		statusCode, errMsg := store.GetStore().UpdateAlert(alert.LastAlertSent)
		if errMsg != "" {
			log.Errorf("failed to update alert for project_id: %v, alert_name: %s, error: %v", projectID, alert.AlertName, errMsg)
			continue
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
	log.Info("query response first", results, statusCode)
	if len(results) != 1 {
		log.Error("empty or invalid result for ", kpiQuery)
		return statusCode, actualValue, comparedValue, errors.New("empty or invalid result")
	}
	if len(results[0].Rows) == 0 {
		log.Error("empty result for ", kpiQuery)
		return statusCode, actualValue, comparedValue, errors.New("empty or invalid value in result")
	}
	if statusCode != http.StatusOK {
		log.Error("failed to execute query for ", kpiQuery)
		return statusCode, actualValue, comparedValue, nil
	}
	if kpiQuery.Category == model.ProfileQueryClass {
		switch results[0].Rows[0][1].(type) {
		case float64:
			actualValue = results[0].Rows[0][1].(float64)
		case int64:
			actualValue = float64(results[0].Rows[0][1].(int64))
		case int:
			actualValue = float64(results[0].Rows[0][1].(int))
		case float32:
			actualValue = float64(results[0].Rows[0][1].(float32))
		default:
			log.Error("failed to convert value to float64 for ", kpiQuery)
			return statusCode, actualValue, comparedValue, errors.New("failed to convert value to float64")
		}
	} else {
		switch results[0].Rows[0][0].(type) {
		case float64:
			actualValue = results[0].Rows[0][0].(float64)
		case int64:
			actualValue = float64(results[0].Rows[0][0].(int64))
		case int:
			actualValue = float64(results[0].Rows[0][0].(int))
		case float32:
			actualValue = float64(results[0].Rows[0][0].(float32))
		default:
			log.Error("invalid value for ", kpiQuery)
			return statusCode, actualValue, comparedValue, errors.New("invalid value")
		}

	}
	if alertType == 2 {
		kpiQueryGroup.Queries = []model.KPIQuery{}
		kpiQuery.From = date_range.prev_from
		kpiQuery.To = date_range.prev_to
		kpiQueryGroup.Queries = append(kpiQueryGroup.Queries, kpiQuery)
		results, statusCode = store.GetStore().ExecuteKPIQueryGroup(projectID, "", kpiQueryGroup)
		log.Info("query response second", results, statusCode)
		if len(results) != 1 {
			log.Error("empty or invalid result for comparision type alerts  ", kpiQuery)
			return statusCode, actualValue, comparedValue, errors.New("empty or invalid result")
		}
		if len(results[0].Rows) == 0 {
			log.Error("empty result for comparision type alerts ", kpiQuery)
			return statusCode, actualValue, comparedValue, errors.New("empty or invalid value in result")
		}
		if statusCode != http.StatusOK {
			return statusCode, actualValue, comparedValue, nil
		}
		if kpiQuery.Category == model.ProfileQueryClass {
			comparedValue = results[0].Rows[0][1].(float64)
		} else {
			comparedValue = results[0].Rows[0][0].(float64)
		}
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
		return false, errors.New("invalid comparsion")
	}
	return false, nil
}

func sendEmailAlert(projectID uint64, msg Message, dateRange dateRanges, timezone U.TimeZoneString, emails []string) {
	var success, fail int
	sub := "Factors Alert"
	text := ""
	var statement string
	fromTime := time.Unix(dateRange.from, 0)
	toTime := time.Unix(dateRange.to, 0)

	fromTime = U.ConvertTimeIn(fromTime, timezone)
	toTime = U.ConvertTimeIn(toTime, timezone)

	from := fromTime.Format("02 Jan 2006")
	to := toTime.Format("02 Jan 2006")

	if msg.Operator == model.INCREASED_OR_DECREASED_BY_MORE_THAN || msg.Operator == model.PERCENTAGE_HAS_INCREASED_OR_DECREASED_BY_MORE_THAN {
		if msg.ActualValue > msg.ComparedValue {
			if msg.Operator == model.INCREASED_OR_DECREASED_BY_MORE_THAN {
				msg.Operator = model.INCREASED_BY_MORE_THAN
			} else {
				msg.Operator = model.PERCENTAGE_HAS_INCREASED_BY_MORE_THAN
			}
		} else {
			if msg.Operator == model.INCREASED_OR_DECREASED_BY_MORE_THAN {
				msg.Operator = model.DECREASED_BY_MORE_THAN
			} else {
				msg.Operator = model.PERCENTAGE_HAS_DECREASED_BY_MORE_THAN
			}
		}
	}
	actualValue := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", msg.ActualValue), "0"), ".")

	if msg.AlertType == 1 {
		if msg.Category == strings.Title(model.PageViews) {
			statement = fmt.Sprintf(`For the %s (%s to %s) <br> <b> %s from %s %s %s : %s </b>`, strings.ReplaceAll(msg.DateRange, "_", " "), from, to, strings.ReplaceAll(msg.AlertName, "_", " "), msg.PageURL, strings.ReplaceAll(msg.Operator, "_", " "), fmt.Sprint(msg.Value), actualValue)
		} else {
			statement = fmt.Sprintf(`For the %s (%s to %s) <br> <b> %s from %s %s %s : %s </b>`, strings.ReplaceAll(msg.DateRange, "_", " "), from, to, strings.ReplaceAll(msg.AlertName, "_", " "), strings.ReplaceAll(msg.Category, "_", " "), strings.ReplaceAll(msg.Operator, "_", " "), fmt.Sprint(msg.Value), actualValue)
		}

		//	statement = fmt.Sprintf(`%s %s recorded for %s in %s from %s to %s`, fmt.Sprint(msg.ActualValue), strings.ReplaceAll(msg.AlertName, "_", " "), strings.ReplaceAll(msg.Category, "_", " "), strings.ReplaceAll(msg.DateRange, "_", " "), from, to)
	} else if msg.AlertType == 2 {
		ComparedValue := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", msg.ComparedValue), "0"), ".")
		if msg.Category == strings.Title(model.PageViews) {
			statement = fmt.Sprintf(`For the %s (%s to %s) compared to %s <br> <b> %s from %s %s %s : %s(%s) </b>`, strings.ReplaceAll(msg.DateRange, "_", " "), from, to, strings.ReplaceAll(msg.ComparedTo, "_", " "), strings.ReplaceAll(msg.AlertName, "_", " "), msg.PageURL, strings.ReplaceAll(msg.Operator, "_", " "), fmt.Sprint(msg.Value), actualValue, ComparedValue)
		} else {
			statement = fmt.Sprintf(`For the %s (%s to %s) compared to %s <br> <b> %s from %s %s %s : %s(%s) </b>`, strings.ReplaceAll(msg.DateRange, "_", " "), from, to, strings.ReplaceAll(msg.ComparedTo, "_", " "), strings.ReplaceAll(msg.AlertName, "_", " "), strings.ReplaceAll(msg.Category, "_", " "), strings.ReplaceAll(msg.Operator, "_", " "), fmt.Sprint(msg.Value), actualValue, ComparedValue)
		}
		//	statement = fmt.Sprintf(`%s %s %s for %s in %s (from %s to %s ) compared to %s - %s(%s)`, strings.ReplaceAll(msg.AlertName, "_", " "), strings.ReplaceAll(msg.Operator, "_", " "), fmt.Sprint(msg.Value), strings.ReplaceAll(msg.Category, "_", " "), strings.ReplaceAll(msg.DateRange, "_", " "), from, to, strings.ReplaceAll(msg.ComparedTo, "_", " "), fmt.Sprint(msg.ActualValue), fmt.Sprint(msg.ComparedValue))
	}
	html := U.CreateAlertTemplate(statement)
	dryRunFlag := C.GetConfig().EnableDryRunAlerts
	if dryRunFlag {
		log.Info("Dry run mode enabled. No emails will be sent")
		log.Info(statement, projectID)
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
	log.Info(statement, projectID)
	log.Info("sent email alert to ", success, " failed to send email alert to ", fail)
}

func sendSlackAlert(msg Message) bool {
	fmt.Println("slack alert sent")
	return true
}
func getGBTForKPIQuery(dateRangeType string) string {
	if dateRangeType == model.LAST_WEEK {
		return model.GroupByTimestampWeek
	} else if dateRangeType == model.LAST_MONTH {
		return model.GroupByTimestampMonth
	} else if dateRangeType == model.LAST_QUARTER {
		return model.GroupByTimestampQuarter
	}
	return ""
}
func filterStringbyLastWord(displayCategory string, word string) string {
	arr := strings.Split(displayCategory, "_")
	// check last element of array is "metrics" if yes delete that
	if arr[len(arr)-1] == word || arr[len(arr)-1] == strings.Title(word) {
		arr = arr[:len(arr)-1]
	}
	return strings.Join(arr, "_")
}
