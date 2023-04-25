package task

import (
	"encoding/json"
	"errors"
	C "factors/config"
	H "factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	slack "factors/integration/slack"
	U "factors/util"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	qc "factors/quickchart"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	AlertName     string
	AlertType     int
	Operator      string
	Category      string
	CategoryType  string
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

const (
	MAX_COLUMNS = 8
	MAX_ROWS    = 10
)

func ComputeAndSendAlerts(projectID int64, configs map[string]interface{}) (map[string]interface{}, bool) {
	allAlerts, errCode := store.GetStore().GetAllAlerts(projectID, true)
	if errCode != http.StatusFound {
		log.Fatalf("Failed to get all alerts for project_id: %v", projectID)
		return nil, false
	}
	status := make(map[string]interface{})
	var alertsToBeProcessed []model.Alert
	if configs["modelType"] == ModelTypeWeek {
		for _, alert := range allAlerts {
			var alertDescription model.AlertDescription
			err := U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
			if err != nil {
				log.Errorf("failed to decode alert description for project_id: %v, alert_name: %s", projectID, alert.AlertName)
				log.Error(err)
				status["error"] = err
				return status, false
			}
			if alertDescription.DateRange == model.LAST_WEEK {
				alertsToBeProcessed = append(alertsToBeProcessed, alert)
			}

		}
	} else if configs["modelType"] == ModelTypeMonth {
		for _, alert := range allAlerts {
			var alertDescription model.AlertDescription
			err := U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
			if err != nil {
				log.Errorf("failed to decode alert description for project_id: %v, alert_name: %s", projectID, alert.AlertName)
				log.Error(err)
				status["error"] = err
				return status, false
			}
			if alertDescription.DateRange == model.LAST_MONTH {
				alertsToBeProcessed = append(alertsToBeProcessed, alert)
			}

		}

	} else if configs["modelType"] == ModelTypeQuarter {
		for _, alert := range allAlerts {
			var alertDescription model.AlertDescription
			err := U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
			if err != nil {
				log.Errorf("failed to decode alert description for project_id: %v, alert_name: %s", projectID, alert.AlertName)
				log.Error(err)
				status["error"] = err
				return status, false
			}
			if alertDescription.DateRange == model.LAST_QUARTER {
				alertsToBeProcessed = append(alertsToBeProcessed, alert)
			}

		}

	} else {
		status["error"] = "invalid model type"
		return status, false
	}
	var alertDescription model.AlertDescription
	var alertConfiguration model.AlertConfiguration
	var dateRange dateRanges

	endTimestampUnix := configs["endTimestamp"].(int64)
	if endTimestampUnix == 0 || endTimestampUnix > U.TimeNowUnix() {
		status["error"] = "invalid end timestamp"
		return status, false
	}
	endTimestamp := time.Unix(endTimestampUnix, 0)
	for _, alert := range alertsToBeProcessed {
		var kpiQuery model.KPIQuery
		alert.LastRunTime = time.Now()
		var err error
		if alert.AlertType == 3 {
			_, err := HandlerAlertWithQueryID(alert, configs, false, 0, 0)
			if err != nil {
				log.WithError(err).Error("failed to run type 3 alert with alert id ", alert.ID)
			}
			continue
		}
		alertDescription, alertConfiguration, kpiQuery, err = model.DecodeAndFetchAlertRelatedStructs(projectID, alert)
		if err != nil {
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
				AlertName:     strings.Title(alert.AlertName),
				AlertType:     alert.AlertType,
				Operator:      alertDescription.Operator,
				Category:      strings.Title(filterStringbyLastWord(kpiQuery.DisplayCategory, "metrics")),
				CategoryType:  model.MapOfMetricsToData[kpiQuery.DisplayCategory][alertDescription.Name]["type"],
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
				sendSlackAlert(alert.ProjectID, alert.CreatedBy, msg, dateRange, timezoneString, alertConfiguration.SlackChannelsAndUserGroups)
			}
		}
		alert.LastAlertSent = true
		statusCode, errMsg := store.GetStore().UpdateAlertStatus(alert.LastAlertSent)
		if errMsg != "" {
			log.Errorf("failed to update alert for project_id: %v, alert_name: %s, error: %v", projectID, alert.AlertName, errMsg)
			continue
		}
	}
	return nil, true
}

func executeAlertsKPIQuery(projectID int64, alertType int, date_range dateRanges,
	kpiQuery model.KPIQuery) (statusCode int, actualValue float64, comparedValue float64, err error) {

	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	kpiQuery.From = date_range.from
	kpiQuery.To = date_range.to
	kpiQueryGroup.Queries = append(kpiQueryGroup.Queries, kpiQuery)
	results, statusCode := store.GetStore().ExecuteKPIQueryGroup(projectID, "",
		kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	//("query response first", results, statusCode)
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
		results, statusCode = store.GetStore().ExecuteKPIQueryGroup(projectID, "",
			kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		//	log.Info("query response second", results, statusCode)
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
			switch results[0].Rows[0][1].(type) {
			case float64:
				comparedValue = results[0].Rows[0][1].(float64)
			case int64:
				comparedValue = float64(results[0].Rows[0][1].(int64))
			case int:
				comparedValue = float64(results[0].Rows[0][1].(int))
			case float32:
				comparedValue = float64(results[0].Rows[0][1].(float32))
			default:
				log.Error("invalid value for ", kpiQuery)
				return statusCode, actualValue, comparedValue, errors.New("invalid value")

			}

		} else {
			switch results[0].Rows[0][0].(type) {
			case float64:
				comparedValue = results[0].Rows[0][0].(float64)
			case int64:
				comparedValue = float64(results[0].Rows[0][0].(int64))
			case int:
				comparedValue = float64(results[0].Rows[0][0].(int))
			case float32:
				comparedValue = float64(results[0].Rows[0][0].(float32))
			default:
				log.Error("invalid value for ", kpiQuery)
				return statusCode, actualValue, comparedValue, errors.New("invalid value")
			}
		}
	}

	return statusCode, actualValue, comparedValue, nil
}

func getDateRange(timezone U.TimeZoneString, dateRange string, prevDateRange string, endTimeStamp time.Time) (dateRanges, error) {
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

func sendEmailAlert(projectID int64, msg Message, dateRange dateRanges, timezone U.TimeZoneString, emails []string) {
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
	percentageSymbol := ""
	if msg.Operator == model.PERCENTAGE_HAS_INCREASED_BY_MORE_THAN || msg.Operator == model.PERCENTAGE_HAS_DECREASED_BY_MORE_THAN {
		percentageSymbol = "%"
		if msg.Operator == model.PERCENTAGE_HAS_INCREASED_BY_MORE_THAN {
			msg.Operator = model.INCREASED_BY_MORE_THAN
		} else {
			msg.Operator = model.DECREASED_BY_MORE_THAN
		}
	}
	var actualValue string
	if msg.CategoryType == model.MetricsDateType {
		actualValue = convertTimeFromSeconds(msg.ActualValue)
	} else {
		actualValue = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", msg.ActualValue), "0"), ".")
		actualValue = AddCommaToNumber(actualValue)
	}

	var comparedValue, comparedToStatement string
	if msg.AlertType == 2 {
		if msg.CategoryType == model.MetricsDateType {
			comparedValue = convertTimeFromSeconds(msg.ComparedValue)
			comparedValue = "(" + comparedValue + ")"
		} else {
			comparedValue = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", msg.ComparedValue), "0"), ".")
			comparedValue = "(" + AddCommaToNumber(comparedValue) + ")"
		}
		previousPeriod := ""
		switch msg.DateRange {
		case model.LAST_WEEK:
			previousPeriod = "week before"
		case model.LAST_MONTH:
			previousPeriod = "month before"
		case model.LAST_QUARTER:
			previousPeriod = "quarter before"
		}
		comparedToStatement = "compared to " + previousPeriod
	}
	// For the last week (10 may 2022 - 16 may 2022) compared to previous period
	// Sessions is increased by more than 100 :  400(250)
	statement = fmt.Sprintf(`For the %s (%s to %s) %s<br> <b> %s %s %s%s : %s%s . </b>`, strings.ReplaceAll(msg.DateRange, "_", " "), from, to, comparedToStatement, strings.ReplaceAll(msg.AlertName, "_", " "), strings.ReplaceAll(msg.Operator, "_", " "), AddCommaToNumber(fmt.Sprint(msg.Value)), percentageSymbol, actualValue, comparedValue)
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
			continue
		}
		success++
	}
	log.Info(statement, projectID)
	log.Info("sent email alert to ", success, " failed to send email alert to ", fail)
}

func sendSlackAlert(projectID int64, agentUUID string, msg Message, dateRange dateRanges, timezone U.TimeZoneString, config *postgres.Jsonb) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"agent_uuid": agentUUID,
	})
	slackChannels := make(map[string][]model.SlackChannel)
	err := U.DecodePostgresJsonbToStructType(config, &slackChannels)
	if err != nil {
		log.WithError(err).Error("failed to decode slack channels")
		return
	}
	var success, fail int
	slackMsg := getSlackMessage(msg, dateRange, timezone)
	dryRunFlag := C.GetConfig().EnableDryRunAlerts
	if dryRunFlag {
		log.Info("Dry run mode enabled. No alerts will be sent")
		log.Info(slackMsg, projectID)
		return
	}
	for _, channels := range slackChannels {
		for _, channel := range channels {
			status, err := slack.SendSlackAlert(projectID, slackMsg, agentUUID, channel)
			if err != nil || !status {
				fail++
				logCtx.WithError(err).Error("failed to send slack alert ", slackMsg)
				continue
			}
			success++
		}
	}
	logCtx.Info("sent slack alert to ", success, " failed to send slack alert to ", fail)

}
func getSlackMessage(msg Message, dateRange dateRanges, timezone U.TimeZoneString) string {
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
	percentageSymbol := ""
	if msg.Operator == model.PERCENTAGE_HAS_INCREASED_BY_MORE_THAN || msg.Operator == model.PERCENTAGE_HAS_DECREASED_BY_MORE_THAN {
		percentageSymbol = "%"
		if msg.Operator == model.PERCENTAGE_HAS_INCREASED_BY_MORE_THAN {
			msg.Operator = model.INCREASED_BY_MORE_THAN
		} else {
			msg.Operator = model.DECREASED_BY_MORE_THAN
		}
	}
	emoji := getEmojiForSlackByOperator(msg.Operator)
	var actualValue string
	if msg.CategoryType == model.MetricsDateType {
		actualValue = convertTimeFromSeconds(msg.ActualValue)
	} else {
		actualValue = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", msg.ActualValue), "0"), ".")
		actualValue = AddCommaToNumber(actualValue)
	}
	var slackMsg string
	comparedToStatement := ""
	ComparedValue := ""
	if msg.AlertType == 2 {
		comparedToStatement = "compared to previous period "
		if msg.CategoryType == model.MetricsDateType {
			ComparedValue = convertTimeFromSeconds(msg.ComparedValue)
		} else {
			ComparedValue = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", msg.ComparedValue), "0"), ".")
			ComparedValue = AddCommaToNumber(ComparedValue)
		}
		ComparedValue = " (" + ComparedValue + ") "
	}
	slackMsg = fmt.Sprintf(`
				[
					{
						"type": "section",
						"text": {
							"type": "mrkdwn",
							"text": "*For the %s (%s - %s) %s *",
						}
					},
					{
						"type": "header",
						"text": {
							"type": "plain_text",
							"text": "%s %s %s%s "
						}
					},
					{
						"type": "divider"
					},
					{
						"type": "section",
						"fields": [
							{
								"type": "mrkdwn",
								"text": "*%s %s*"
							},
							{
								"type": "mrkdwn",
								"text": "%s"
							}
						]
					},
					{
						"type": "divider"
					},
					{
						"type": "section",
						"text": {
							"type": "mrkdwn",
							"text": "*<https://app.factors.ai/|Go to Factors.ai>*"
						}
					},
					{
						"type": "divider"
					}
				]
				`, strings.ReplaceAll(msg.DateRange, "_", " "), from, to, comparedToStatement, strings.ReplaceAll(msg.AlertName, "_", " "), strings.ReplaceAll(msg.Operator, "_", " "), AddCommaToNumber(fmt.Sprint(msg.Value)), percentageSymbol, actualValue, ComparedValue, emoji)

	return slackMsg
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
func HandlerAlertWithQueryID(alert model.Alert, configs map[string]interface{}, shouldOverride bool, overRideFrom, overRideTo int64) (bool, error) {
	logCtx := log.WithFields(log.Fields{
		"projectID": alert.ProjectID,
		"alertID":   alert.ID,
	})
	projectID := alert.ProjectID
	query, status := store.GetStore().GetQueryWithQueryId(projectID, alert.QueryID)
	if status != http.StatusFound {
		log.WithContext(logCtx.Context).Error("Query not found for id ", alert.QueryID)
		return false, errors.New("Query not found")
	}
	class, errMsg := store.GetStore().GetQueryClassFromQueries(*query)
	if errMsg != "" {
		log.Error("Class not Found for queryID ", errMsg)
		return false, errors.New(errMsg)
	}
	var alertConfiguration model.AlertConfiguration
	err := U.DecodePostgresJsonbToStructType(alert.AlertConfiguration, &alertConfiguration)
	if err != nil {
		log.Errorf("failed to decode alert configuration for project_id: %v, alert_name: %s", projectID, alert.AlertName)
		log.Error(err)
		return false, errors.New(fmt.Sprintf("failed to decode alert configuration for project_id: %v, alert_name: %s", projectID, alert.AlertName))
	}
	if class == model.QueryClassEvents {
		_, err = handleShareQueryTypeEvents(alert, query, configs, shouldOverride, overRideFrom, overRideTo)
		if err != nil {
			logCtx.Error(err)
		}
		return true, nil
	} else if class == model.QueryClassKPI {
		_, err = handleShareQueryTypeKPI(alert, query, configs, shouldOverride, overRideFrom, overRideTo)
		if err != nil {
			logCtx.Error(err)
		}
		return true, nil
	}
	return false, errors.New("query type not supported for this operation")
}
func handleShareQueryTypeEvents(alert model.Alert, query *model.Queries, configs map[string]interface{}, shouldOverride bool, overRideFrom, overRideTo int64) (bool, error) {
	projectID := alert.ProjectID
	queryGroup := model.QueryGroup{}
	U.DecodePostgresJsonbToStructType(&query.Query, &queryGroup)
	if len(queryGroup.Queries) == 0 {
		log.Error("Query failed. Empty query group.")
		return false, errors.New("Query failed. Empty query group.")
	}
	var alertConfiguration model.AlertConfiguration
	err := U.DecodePostgresJsonbToStructType(alert.AlertConfiguration, &alertConfiguration)
	if err != nil {
		log.Errorf("failed to decode alert configuration for project_id: %v, alert_name: %s", projectID, alert.AlertName)
		log.Error(err)
		return false, errors.New(fmt.Sprintf("failed to decode alert configuration for project_id: %v, alert_name: %s", projectID, alert.AlertName))
	}
	containsBreakdown := false
	noOfbreakdowns := 0
	for _, query := range queryGroup.Queries {
		if query.GroupByProperties != nil && len(query.GroupByProperties) > 0 {
			containsBreakdown = true
			noOfbreakdowns = len(queryGroup.Queries[0].GroupByProperties)
		}
	}

	var alertDescription model.AlertDescription
	err = U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
	if err != nil {
		log.Errorf("failed to decode alert description for project_id: %v, alert_name: %s", projectID, alert.AlertName)
		log.Error(err)
		return false, err
	}

	var timezoneString U.TimeZoneString
	var statusCode int
	if queryGroup.Queries[0].Timezone != "" {
		_, errCode := time.LoadLocation(string(queryGroup.Queries[0].Timezone))
		if errCode != nil {
			log.Error("Query failed. Invalid timezone.")
			return false, errors.New("Query failed. Invalid timezone.")
		}
		timezoneString = U.TimeZoneString(queryGroup.Queries[0].Timezone)
	} else {
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			log.Error("Query failed. Invalid timezone.")
			return false, errors.New("Query failed. Invalid timezone.")
		}
		// logCtx.WithError(err).Error("Query failed. Invalid Timezone.")
	}
	queryGroup.SetTimeZone(timezoneString)

	err = queryGroup.TransformDateTypeFilters()
	if err != nil {
		log.Error("Query failed. Error in date type filter.")
		return false, err

	}
	var endTimeStamp time.Time
	if configs != nil {
		endTimestampUnix := configs["endTimestamp"].(int64)
		if endTimestampUnix == 0 || endTimestampUnix > U.TimeNowUnix() {
			return false, errors.New("invalid end time stamp")
		}
		endTimeStamp = time.Unix(endTimestampUnix, 0)
	}
	title := strings.Title(alert.AlertName)

	var dateRange dateRanges

	if alertDescription.DateRange != "" {
		dateRange, err = getDateRange(timezoneString, alertDescription.DateRange, "", endTimeStamp)
		for _, query := range queryGroup.Queries {
			query.From = dateRange.from
			query.To = dateRange.to
		}
	}

	if shouldOverride {
		for idx := range queryGroup.Queries {
			queryGroup.Queries[idx].From = overRideFrom
			queryGroup.Queries[idx].To = overRideTo
		}
	}

	from := time.Unix(queryGroup.Queries[0].From, 0)
	to := time.Unix(queryGroup.Queries[0].To, 0)
	fromTime := U.ConvertTimeIn(from, timezoneString)
	toTime := U.ConvertTimeIn(to, timezoneString)
	fromStr := fromTime.Format("02 Jan 2006")
	toStr := toTime.Format("02 Jan 2006")
	newDateRange := fromStr + " - " + toStr
	var cacheResult model.ResultGroup
	// fix is Dashboard query request flag
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(nil, projectID, &queryGroup, cacheResult, false, "", true)
	if shouldReturn {
		if resCode == http.StatusOK {
			// trigger the alert based on resMsg
			tempjson, _ := json.Marshal(resMsg)
			err := json.Unmarshal(tempjson, &cacheResult)
			if err != nil {
				log.Error(err)
				return false, err
			}
			if alertConfiguration.IsEmailEnabled || alertConfiguration.IsSlackEnabled {
				chartUrl, tableUrl, skipChart, err := getUrlsForSavedQuerySharing(alert, model.QueryClassEvents, title, newDateRange, containsBreakdown, noOfbreakdowns, cacheResult.Results)
				if err != nil {
					log.WithError(err).Error("Failed to build charts and table urls")
					return false, errors.New("Failed to build charts and table urls")
				}
				if alertConfiguration.IsEmailEnabled {
					//tableRows := getEmailTemplateForSavedQuerySharing(alert, model.QueryClassKPI, queryResult)
					sendEmailSavedReport(title, alertDescription.Subject, newDateRange, chartUrl, tableUrl, alertConfiguration.Emails)
				}
				if alertConfiguration.IsSlackEnabled {
					sendSlackAlertForSavedQueries(alert, newDateRange, model.QueryClassEvents, title, chartUrl, tableUrl, skipChart, alertConfiguration.SlackChannelsAndUserGroups)
				}
			}
		}
		log.Error("Query failed. Error Processing/Fetching data from Query cache")
		return false, errors.New("Query failed. Error Processing/Fetching data from Query cache")
	}
	resultGroup, errCode := store.GetStore().RunEventsGroupQuery(queryGroup.Queries, alert.ProjectID, C.EnableOptimisedFilterOnEventUserQuery())
	if errCode != http.StatusOK {
		model.DeleteQueryCacheKey(alert.ProjectID, &queryGroup)
		log.Error("Query failed. Failed to process query from DB")
		if errCode == http.StatusPartialContent {
			return false, errors.New("Query failed. Failed to process query from DB .")
		}
		return false, errors.New("Query failed. Failed to process query from DB")
	}

	if alertConfiguration.IsEmailEnabled || alertConfiguration.IsSlackEnabled {
		chartUrl, tableUrl, skipChart, err := getUrlsForSavedQuerySharing(alert, model.QueryClassEvents, title, newDateRange, containsBreakdown, noOfbreakdowns, resultGroup.Results)
		if err != nil {
			log.WithError(err).Error("Failed to build charts and table urls")
			return false, errors.New("Failed to build charts and table urls")
		}
		if alertConfiguration.IsEmailEnabled {
			//tableRows := getEmailTemplateForSavedQuerySharing(alert, model.QueryClassKPI, queryResult)
			sendEmailSavedReport(title, alertDescription.Subject, newDateRange, chartUrl, tableUrl, alertConfiguration.Emails)
		}
		if alertConfiguration.IsSlackEnabled {
			sendSlackAlertForSavedQueries(alert, newDateRange, model.QueryClassEvents, title, chartUrl, tableUrl, skipChart, alertConfiguration.SlackChannelsAndUserGroups)
		}
	}
	return true, nil
}
func handleShareQueryTypeKPI(alert model.Alert, query *model.Queries, configs map[string]interface{}, shouldOverride bool, overRideFrom, overRideTo int64) (bool, error) {
	projectID := alert.ProjectID
	kpiQueryGroup := model.KPIQueryGroup{}
	U.DecodePostgresJsonbToStructType(&query.Query, &kpiQueryGroup)
	if len(kpiQueryGroup.Queries) == 0 {
		log.Error("Query failed. Empty query group.")
		return false, errors.New("Query failed. Empty query group.")
	}
	noOfBreakdowns := 0
	containsBreakdown := false
	if kpiQueryGroup.GlobalGroupBy != nil && len(kpiQueryGroup.GlobalGroupBy) > 0 {
		containsBreakdown = true
		noOfBreakdowns = len(kpiQueryGroup.GlobalGroupBy)
	}
	var alertConfiguration model.AlertConfiguration
	err := U.DecodePostgresJsonbToStructType(alert.AlertConfiguration, &alertConfiguration)
	if err != nil {
		log.Errorf("failed to decode alert configuration for project_id: %v, alert_name: %s", projectID, alert.AlertName)
		log.Error(err)
		return false, errors.New(fmt.Sprintf("failed to decode alert configuration for project_id: %v, alert_name: %s", projectID, alert.AlertName))
	}
	var alertDescription model.AlertDescription
	err = U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
	if err != nil {
		log.Errorf("failed to decode alert description for project_id: %v, alert_name: %s", projectID, alert.AlertName)
		log.Error(err)
		return false, err
	}
	var timezoneString U.TimeZoneString
	var statusCode int
	if kpiQueryGroup.Queries[0].Timezone != "" {
		_, errCode := time.LoadLocation(string(kpiQueryGroup.Queries[0].Timezone))
		if errCode != nil {
			return false, errors.New("Query failed. Invalid Timezone provided.")
		}

		timezoneString = U.TimeZoneString(kpiQueryGroup.Queries[0].Timezone)
	} else {
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			log.Error("Query failed. Failed to get Timezone.")
			return false, errors.New("Query failed. Failed to get Timezone.")
		}
		// logCtx.WithError(err).Error("Query failed. Invalid Timezone.")
	}
	kpiQueryGroup.SetTimeZone(timezoneString)
	err = kpiQueryGroup.TransformDateTypeFilters()
	if err != nil {
		log.Error("Query failed. Error in date type filter.")
		return false, err

	}
	var endTimeStamp time.Time
	if configs != nil {
		endTimestampUnix := configs["endTimestamp"].(int64)
		if endTimestampUnix == 0 || endTimestampUnix > U.TimeNowUnix() {
			return false, errors.New("invalid end time stamp")
		}
		endTimeStamp = time.Unix(endTimestampUnix, 0)
	}
	title := alert.AlertName

	var dateRange dateRanges

	if alertDescription.DateRange != "" {
		dateRange, err = getDateRange(timezoneString, alertDescription.DateRange, "", endTimeStamp)
		for _, query := range kpiQueryGroup.Queries {
			query.From = dateRange.from
			query.To = dateRange.to
		}
	}

	if shouldOverride {
		for idx := range kpiQueryGroup.Queries {
			kpiQueryGroup.Queries[idx].From = overRideFrom
			kpiQueryGroup.Queries[idx].To = overRideTo
		}
	}
	from := time.Unix(kpiQueryGroup.Queries[0].From, 0)
	to := time.Unix(kpiQueryGroup.Queries[0].To, 0)
	fromTime := U.ConvertTimeIn(from, timezoneString)
	toTime := U.ConvertTimeIn(to, timezoneString)
	fromStr := fromTime.Format("02 Jan 2006")
	toStr := toTime.Format("02 Jan 2006")
	newDateRange := fromStr + " - " + toStr

	var cacheResult []model.QueryResult
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(nil, projectID, &kpiQueryGroup, cacheResult, false, "", true)
	if shouldReturn {
		if resCode == http.StatusOK {
			// trigger the alert based on resMsg
			tempjson, _ := json.Marshal(resMsg)
			err := json.Unmarshal(tempjson, &cacheResult)
			if err != nil {
				log.Error(err)
				return false, err
			}
			if alertConfiguration.IsEmailEnabled || alertConfiguration.IsSlackEnabled {
				chartUrl, tableUrl, skipChart, err := getUrlsForSavedQuerySharing(alert, model.QueryClassKPI, title, newDateRange, containsBreakdown, noOfBreakdowns, cacheResult)
				if err != nil {
					log.WithError(err).Error("Failed to build charts and table urls")
					return false, errors.New("Failed to build charts and table urls")
				}
				if alertConfiguration.IsEmailEnabled {
					//tableRows := getEmailTemplateForSavedQuerySharing(alert, model.QueryClassKPI, queryResult)
					sendEmailSavedReport(title, alertDescription.Subject, newDateRange, chartUrl, tableUrl, alertConfiguration.Emails)
				}
				if alertConfiguration.IsSlackEnabled {
					sendSlackAlertForSavedQueries(alert, newDateRange, model.QueryClassKPI, title, chartUrl, tableUrl, skipChart, alertConfiguration.SlackChannelsAndUserGroups)
				}
			}
			return true, nil
		}
		log.Error("Query failed. Error Processing/Fetching data from Query cache")
		return false, errors.New("Query failed. Error Processing/Fetching data from Query cache")
	}
	var duplicatedRequest model.KPIQueryGroup
	U.DeepCopy(&kpiQueryGroup, &duplicatedRequest)
	queryResult, statusCode := store.GetStore().ExecuteKPIQueryGroup(projectID, kpiQueryGroup.Class, duplicatedRequest, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	if statusCode != http.StatusOK {
		model.DeleteQueryCacheKey(projectID, &kpiQueryGroup)
		log.Error("Failed to process query from DB")
		if statusCode == http.StatusPartialContent {
			return false, errors.New("Failed to process query from DB .")
		}
		return false, errors.New("Failed to process query from DB")
	}
	if alertConfiguration.IsEmailEnabled || alertConfiguration.IsSlackEnabled {
		chartUrl, tableUrl, skipChart, err := getUrlsForSavedQuerySharing(alert, model.QueryClassKPI, title, newDateRange, containsBreakdown, noOfBreakdowns, queryResult)
		if err != nil {
			log.WithError(err).Error("Failed to build charts and table urls")
			return false, errors.New("Failed to build charts and table urls")
		}
		if alertConfiguration.IsEmailEnabled {
			//tableRows := getEmailTemplateForSavedQuerySharing(alert, model.QueryClassKPI, queryResult)
			sendEmailSavedReport(title, alertDescription.Subject, newDateRange, chartUrl, tableUrl, alertConfiguration.Emails)
		}
		if alertConfiguration.IsSlackEnabled {
			sendSlackAlertForSavedQueries(alert, newDateRange, model.QueryClassKPI, title, chartUrl, tableUrl, skipChart, alertConfiguration.SlackChannelsAndUserGroups)
		}
	}
	return true, nil

}
func getUrlsForSavedQuerySharing(alert model.Alert, queryClass, reportTitle, dateRange string, containsBreakdown bool, noOfBreakdowns int, result []model.QueryResult) (string, string, bool, error) {
	tableConfig := buildTableConfigForSavedQuerySharing(alert, queryClass, reportTitle, dateRange, containsBreakdown, result)

	tableUrl, err := qc.GetTableURLfromTableConfig(tableConfig)
	if err != nil {
		log.WithError(err).Error("Failed to get table url from table config")
		return "", "", false, err
	}
	var chartUrl string
	skipChart := queryClass == model.QueryClassEvents && containsBreakdown
	if !skipChart {
		_, displayNames := store.GetStore().GetDisplayNamesForAllEvents(alert.ProjectID)
		displayNameEvents := GetDisplayEventNamesHandler(displayNames)
		chartConfig := buildChartConfigForSavedQuerySharing(queryClass, result, containsBreakdown, noOfBreakdowns, displayNameEvents)
		chartUrl, err = qc.GetChartImageUrlForConfig(chartConfig)
		if err != nil {
			log.WithError(err).Error("Failed to get chart url from chart config")
			return "", "", false, err
		}
	}
	return chartUrl, tableUrl, skipChart, nil
}

// old
func sendEmailSavedQueryReport(reportTitle, subject, date, tableRowsemails string, emails []string) {
	sub := subject
	text := ""
	var success, fail int

	html := U.GetSavedQuerySharingEmailTemplate(reportTitle, date, tableRowsemails)
	dryRunFlag := C.GetConfig().EnableDryRunAlerts
	if dryRunFlag {
		log.Info("Dry run mode enabled. No emails will be sent")
		log.Info(tableRowsemails)
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

// new
func sendEmailSavedReport(reportTitle, subject, date, chartUrl, tableUrl string, emails []string) {
	sub := subject
	text := ""
	var success, fail int

	html := U.GetSavedReportSharingEmailTemplate(reportTitle, date, chartUrl, tableUrl)
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
func getEmailTemplateForSavedQuerySharing(alert model.Alert, queryClass string, result []model.QueryResult) (tableRows string) {
	tableRows = ""
	_, displayNames := store.GetStore().GetDisplayNamesForAllEvents(alert.ProjectID)
	displayNameEvents := GetDisplayEventNamesHandler(displayNames)
	if queryClass == model.QueryClassEvents {
		for _, row := range result[1].Rows {
			tempTitle := row[1].(string)
			if _, exists := displayNameEvents[tempTitle]; exists {
				tempTitle = displayNameEvents[tempTitle]
			}
			title := strings.Title(strings.ReplaceAll(tempTitle, "_", " "))
			var resultValue float64
			switch row[2].(type) {
			case float64:
				resultValue = row[2].(float64)
			case int64:
				resultValue = float64(row[2].(int64))
			case int:
				resultValue = float64(row[2].(int))
			case float32:
				resultValue = float64(row[2].(float32))
			}
			value := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", resultValue), "0"), ".")
			value = AddCommaToNumber(fmt.Sprint(value))
			tableRow := getTableRowForEmailSharing(title, value)
			tableRows += tableRow
		}
	} else if queryClass == model.QueryClassKPI {
		for idx, row := range result[1].Rows[0] {
			tempTitle := result[1].Headers[idx]
			if _, exists := displayNameEvents[tempTitle]; exists {
				tempTitle = displayNameEvents[tempTitle]
			}
			title := strings.Title(strings.ReplaceAll(tempTitle, "_", " "))
			var resultValue float64
			switch row.(type) {
			case float64:
				resultValue = row.(float64)
			case int64:
				resultValue = float64(row.(int64))
			case int:
				resultValue = float64(row.(int))
			case float32:
				resultValue = float64(row.(float32))
			}
			value := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", resultValue), "0"), ".")
			value = AddCommaToNumber(fmt.Sprint(value))
			tableRow := getTableRowForEmailSharing(title, value)
			tableRows += tableRow
		}
	}
	return tableRows
}

func getTableRowForEmailSharing(title, result string) string {
	title = strings.Title(title)
	title = removeDollarSymbolFromEventNames(title)
	tableRow := fmt.Sprintf(`%s</p><h2 class="text-center h2 text-bold float-center" align="center" style="Margin:0;Margin-bottom:10px;color:inherit;font-family:'Open Sans',sans-serif;font-size:30px;font-weight:700;line-height:1.5;margin:0;margin-bottom:10px;padding-bottom:0;padding-left:0;padding-right:0;padding-top:0;text-align:center;word-wrap:normal">%s</h2><p class="text-center h7 medium-grey float-center" align="center" style="Margin:0;Margin-bottom:10px;color:#8692A3;font-family:'Open Sans',sans-serif;font-size:14px;font-weight:400;line-height:1.5;margin:0;margin-bottom:10px;padding-bottom:0;padding-left:0;padding-right:0;padding-top:0;text-align:center"></p><table align="center" class="row float-center" style="Margin:0 auto;border-collapse:collapse;border-spacing:0;display:table;float:none;margin:0 auto;padding:0;padding-bottom:0;padding-left:0;padding-right:0;padding-top:0;position:relative;text-align:center;vertical-align:top;width:100%%"><tbody><tr style="padding-bottom:0;padding-left:0;padding-right:0;padding-top:0;text-align:left;vertical-align:top"><th class="divider-line small-12 large-4 columns first last" style="-moz-hyphens:auto;-webkit-hyphens:auto;Margin:0 auto;border-collapse:collapse!important;border-top:1px solid #E7E9ED;color:#3E516C;display:block;font-family:'Open Sans',sans-serif;font-size:16px;font-weight:400;height:1px;hyphens:auto;line-height:1.5;margin:0 auto;padding:5px 0;padding-bottom:16px;padding-left:0!important;padding-right:0!important;padding-top:0;text-align:left;vertical-align:top;width:33.33333%%;word-wrap:break-word"><table style="border-collapse:collapse;border-spacing:0;padding-bottom:0;padding-left:0;padding-right:0;padding-top:0;text-align:left;vertical-align:top;width:100%%"><tbody><tr style="padding-bottom:0;padding-left:0;padding-right:0;padding-top:0;text-align:left;vertical-align:top"><th style="-moz-hyphens:auto;-webkit-hyphens:auto;Margin:0;border-collapse:collapse!important;color:#3E516C;font-family:'Open Sans',sans-serif;font-size:16px;font-weight:400;hyphens:auto;line-height:1.5;margin:0;padding-bottom:0;padding-left:0;padding-right:0;padding-top:0;text-align:left;vertical-align:top;word-wrap:break-word"></th></tr></tbody></table></th></tr></tbody></table></center><center style="min-width:338.67px;width:100%%"><p class="text-center h7 float-center" align="center" style="Margin:0;Margin-bottom:10px;color:#3E516C;font-family:'Open Sans',sans-serif;font-size:14px;font-weight:400;line-height:1.5;margin:0;margin-bottom:10px;padding-bottom:0;padding-left:0;padding-right:0;padding-top:0;text-align:center">`, title, result)
	return tableRow
}

func sendSlackAlertForSavedQueries(alert model.Alert, dateRange, queryClass, reportTitle, chartUrl, tableUrl string, skipChart bool, config *postgres.Jsonb) {
	logCtx := log.WithFields(log.Fields{
		"project_id": alert.ProjectID,
	})
	slackChannels := make(map[string][]model.SlackChannel)
	err := U.DecodePostgresJsonbToStructType(config, &slackChannels)
	if err != nil {
		log.WithError(err).Error("failed to decode slack channels")
		return
	}
	var success, fail int
	//slackMsg := buildSlackTemplateForSavedQuerySharing(alert, queryClass, reportTitle, dateRange, result)

	slackMsg := buildSlackMessageForReportSharing(queryClass, reportTitle, dateRange, chartUrl, tableUrl, skipChart)
	dryRunFlag := C.GetConfig().EnableDryRunAlerts
	if dryRunFlag {
		log.Info("Dry run mode enabled. No alerts will be sent")
		log.Info(slackMsg, alert.ProjectID)
		return
	}
	for _, channels := range slackChannels {
		for _, channel := range channels {
			status, err := slack.SendSlackAlert(alert.ProjectID, slackMsg, alert.CreatedBy, channel)
			if err != nil || !status {
				fail++
				logCtx.WithError(err).Error("failed to send slack alert ", slackMsg)
				continue
			}
			success++
		}
	}
	logCtx.Info("sent slack alert to ", success, " failed to send slack alert to ", fail)

}

func buildSlackTemplateForSavedQuerySharing(alert model.Alert, queryClass, reportTitle, dateRange string, result []model.QueryResult) string {
	slackBlocks := ""
	_, displayNames := store.GetStore().GetDisplayNamesForAllEvents(alert.ProjectID)
	displayNameEvents := GetDisplayEventNamesHandler(displayNames)
	if queryClass == model.QueryClassEvents {
		for _, row := range result[1].Rows {
			tempTitle := row[1].(string)
			if _, exists := displayNameEvents[tempTitle]; exists {
				tempTitle = displayNameEvents[tempTitle]
			}
			title := strings.Title(strings.ReplaceAll(tempTitle, "_", " "))
			var resultValue float64
			switch row[2].(type) {
			case float64:
				resultValue = row[2].(float64)
			case int64:
				resultValue = float64(row[2].(int64))
			case int:
				resultValue = float64(row[2].(int))
			case float32:
				resultValue = float64(row[2].(float32))
			}
			value := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", resultValue), "0"), ".")
			value = AddCommaToNumber(fmt.Sprint(resultValue))
			block := getSlackBlockForResult(title, value)
			slackBlocks += block
		}
	} else if queryClass == model.QueryClassKPI {
		for idx, row := range result[1].Rows[0] {
			tempTitle := result[1].Headers[idx]
			if _, exists := displayNameEvents[tempTitle]; exists {
				tempTitle = displayNameEvents[tempTitle]
			}
			title := strings.Title(tempTitle)
			var resultValue float64
			switch row.(type) {
			case float64:
				resultValue = row.(float64)
			case int64:
				resultValue = float64(row.(int64))
			case int:
				resultValue = float64(row.(int))
			case float32:
				resultValue = float64(row.(float32))
			}
			value := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", resultValue), "0"), ".")
			value = AddCommaToNumber(fmt.Sprint(value))
			block := getSlackBlockForResult(title, value)
			slackBlocks += block
		}
	}

	slackTemplate := fmt.Sprintf(`
		 [
			{
				"type": "header",
				"text": {
					"type": "plain_text",
					"text": "%s\n"
				}
			},
			{
				"type": "context",
				"elements": [
					{
						"text": "*%s*  ",
						"type": "mrkdwn"
					}
				]
			}
			%s,
			{
				"type": "divider"
			}
		]
	`, reportTitle, dateRange, slackBlocks)

	return slackTemplate
}
func buildTableConfigForSavedQuerySharing(alert model.Alert, queryClass, reportTitle, dateRange string, containsBreakdown bool, result []model.QueryResult) qc.TableConfig {
	var config qc.TableConfig
	var columns []qc.Column
	currRow := 0
	// limiting rows to max rows allowed
	currCol := 0
	// limiting cols to max rows allowed
	DataSource := []interface{}{
		"-",
	}
	_, displayNames := store.GetStore().GetDisplayNamesForAllEvents(alert.ProjectID)
	displayNameEvents := GetDisplayEventNamesHandler(displayNames)
	ColumnMapToDataIndex := make(map[int]string)
	if queryClass == model.QueryClassEvents {
		if containsBreakdown {
			// // sorting based on event selected index
			// sort.Slice(result[1].Rows, func(i, j int) bool {
			// 	return result[1].Rows[i][0].(int) < result[1].Rows[j][0].(int)
			// })

			for idx, header := range result[1].Headers {
				if currCol >= MAX_COLUMNS {
					break
				}
				var column qc.Column
				// ignoring the 0th as its event index
				if idx > 0 {
					tempTitle := header
					if _, exists := displayNameEvents[tempTitle]; exists {
						tempTitle = displayNameEvents[tempTitle]
					} else {
						tempTitle = strings.Title(strings.ReplaceAll(tempTitle, "_", " "))

					}
					tempTitle = removeDollarSymbolFromEventNames(tempTitle)
					column.Title = tempTitle
					column.DataIndex = strings.ToLower(tempTitle) + strconv.Itoa(idx)
					column.Width = len(column.Title) * 10
					ColumnMapToDataIndex[idx] = column.DataIndex
					columns = append(columns, column)
					currCol++
				}
			}
			for _, rows := range result[1].Rows {
				if currRow >= MAX_ROWS {
					break
				}
				dataSourcetemp := map[string]interface{}{}
				for idx, row := range rows {
					if idx > 0 {
						key := ColumnMapToDataIndex[idx]
						value := row
						flag := true
						switch value.(type) {
						case string:
							if _, exists := displayNameEvents[value.(string)]; exists {
								value = displayNameEvents[value.(string)]
							}
							flag = false
						case float64:
							value = (float64(value.(float64)))
						case float32:
							value = (float64(value.(float32)))
						case int:
							value = (float64(value.(int)))
						case int64:
							value = (float64(value.(int64)))
						}
						if flag {
							value = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
							value = AddCommaToNumber(fmt.Sprint(value))
						}
						dataSourcetemp[key] = value
					}
				}
				DataSource = append(DataSource, dataSourcetemp)
				currRow++
			}
		} else {
			for idx, header := range result[0].Headers {
				if currCol >= MAX_COLUMNS {
					break
				}
				var column qc.Column
				tempTitle := header
				if _, exists := displayNameEvents[header]; exists {
					tempTitle = displayNameEvents[header]
				} else {
					tempTitle = strings.Title(strings.ReplaceAll(tempTitle, "_", " "))
				}
				tempTitle = removeDollarSymbolFromEventNames(tempTitle)
				column.Title = tempTitle
				column.DataIndex = strings.ToLower(tempTitle) + strconv.Itoa(idx)
				column.Width = len(column.Title) * 10
				ColumnMapToDataIndex[idx] = column.DataIndex
				columns = append(columns, column)
				currCol++

			}
			for _, rows := range result[0].Rows {
				if currRow >= MAX_ROWS {
					break
				}
				dataSourcetemp := map[string]interface{}{}
				for idx, row := range rows {
					key := ColumnMapToDataIndex[idx]
					value := row
					flag := true
					switch value.(type) {
					case string:
						if _, exists := displayNameEvents[value.(string)]; exists {
							value = displayNameEvents[value.(string)]
						}
						flag = false
					case float64:
						value = (float64(value.(float64)))
					case float32:
						value = (float64(value.(float32)))
					case int:
						value = (float64(value.(int)))
					case int64:
						value = (float64(value.(int64)))
					}
					if flag{
						value = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
						value = AddCommaToNumber(fmt.Sprint(value))
					}

					dataSourcetemp[key] = value

				}
				DataSource = append(DataSource, dataSourcetemp)
				currRow++
			}
		}
	} else if queryClass == model.QueryClassKPI {
		// building columns
		resultIdx := 0
		if containsBreakdown {
			resultIdx = 1
		}
		for idx, header := range result[resultIdx].Headers {
			if currCol >= MAX_COLUMNS {
				break
			}
			var column qc.Column
			tempTitle := header
			if _, exists := displayNameEvents[tempTitle]; exists {
				tempTitle = displayNameEvents[tempTitle]
			} else {
				tempTitle = strings.Title(strings.ReplaceAll(tempTitle, "_", " "))
			}
			tempTitle = removeDollarSymbolFromEventNames(tempTitle)
			column.Title = tempTitle
			column.DataIndex = strings.ToLower(tempTitle) + strconv.Itoa(idx)
			column.Width = len(column.Title) * 10
			ColumnMapToDataIndex[idx] = column.DataIndex
			columns = append(columns, column)
			currCol++

		}
		for _, rows := range result[resultIdx].Rows {
			if currRow >= MAX_ROWS {
				break
			}
			dataSourcetemp := map[string]interface{}{}
			for idx, row := range rows {
				key := ColumnMapToDataIndex[idx]
				value := row
				flag:= true
				switch value.(type) {
				case string:
					if _, exists := displayNameEvents[value.(string)]; exists {
						value = displayNameEvents[value.(string)]
					}
					flag = false
				case float64:
					value = (float64(value.(float64)))
				case float32:
					value = (float64(value.(float32)))
				case int:
					value = (float64(value.(int)))
				case int64:
					value = (float64(value.(int64)))
				}
				if flag {
					value = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
					value = AddCommaToNumber(fmt.Sprint(value))
				}
				dataSourcetemp[key] = value
			}
			DataSource = append(DataSource, dataSourcetemp)
			currRow++
		}

	}
	DataSource = append(DataSource, "-")
	config.DataSource = DataSource
	config.Columns = columns
	return config

}
func buildChartConfigForSavedQuerySharing(queryClass string, results []model.QueryResult, containsBreakDown bool, noOfBreakdowns int, displayNameEvents map[string]string) qc.ChartConfig {
	var config qc.ChartConfig
	if containsBreakDown {
		config.Type = "bar"
		if queryClass == model.QueryClassKPI {
			var chartData qc.ChartData
			var dataSet qc.Dataset
			dataSet.Label = results[1].Headers[noOfBreakdowns]
			if _, exists := displayNameEvents[dataSet.Label]; exists {
				dataSet.Label = displayNameEvents[dataSet.Label]
			}
			dataSet.Label = strings.Title(strings.ReplaceAll(dataSet.Label, "_", " "))
			for _, rows := range results[1].Rows {
				label := ""
				idx := 0
				for idx < noOfBreakdowns {
					label += rows[idx].(string)
					if idx != noOfBreakdowns-1 {
						label += ", "
					}
					idx++
				}
				dataSet.Data = append(dataSet.Data, rows[noOfBreakdowns])
				chartData.Labels = append(chartData.Labels, label)
			}
			chartData.DataSets = append(chartData.DataSets, dataSet)
			config.Data = chartData

		} else if queryClass == model.QueryClassEvents {
			// charts are not supported for events queries with breakdown V1
		}

	} else {
		config.Type = "line"
		if queryClass == model.QueryClassKPI {
			var chartData qc.ChartData
			var dataSet []qc.Dataset
			dataSet = make([]qc.Dataset, len(results[1].Headers))
			for idx, header := range results[1].Headers {
				if _, exists := displayNameEvents[header]; exists {
					header = displayNameEvents[header]
				}
				dataSet[idx].Label = header
			}
			for _, rows := range results[0].Rows {
				for idx, row := range rows {
					if idx == 0 {
						chartData.Labels = append(chartData.Labels, row)
					} else {
						dataSet[idx-1].Data = append(dataSet[idx-1].Data, row)
						dataSet[idx-1].LineTension = 0.4
					}
				}
			}
			chartData.DataSets = dataSet
			config.Data = chartData

		} else if queryClass == model.QueryClassEvents {
			var chartData qc.ChartData
			var dataSet []qc.Dataset
			dataSet = make([]qc.Dataset, len(results[0].Headers)-1)
			for idx, header := range results[0].Headers {
				if idx > 0 {
					if _, exists := displayNameEvents[header]; exists {
						header = displayNameEvents[header]
					}
					dataSet[idx-1].Label = header
				}
			}
			for _, rows := range results[0].Rows {
				for idx, row := range rows {
					if idx == 0 {
						chartData.Labels = append(chartData.Labels, row)
					} else {
						dataSet[idx-1].Data = append(dataSet[idx-1].Data, row)
						dataSet[idx-1].LineTension = 0.4
					}
				}
			}
			chartData.DataSets = dataSet
			config.Data = chartData
		}

	}

	return config
}
func buildSlackMessageForReportSharing(queryClass, reportTitle, dateRange, chartUrl, tableUrl string, eventsWithBreadkown bool) string {
	var slackTemplate string
	if eventsWithBreadkown {
		slackTemplate = fmt.Sprintf(`
		[
				{
					"type": "header",
					"text": {
						"type": "plain_text",
						"text": "%s\n"
					}
				},
				{
					"type": "context",
					"elements": [
						{
							"type": "mrkdwn",
							"text": "*%s*"
						}
					]
				},
				{
					"type": "divider"
				},
				{
					"type": "divider"
				},
				{
					"type": "image",
					"image_url": "%s",
					"alt_text": ""
				},
				 {
					"type": "divider"
				}
		]
	
		`, reportTitle, dateRange, tableUrl)
	} else {
		slackTemplate = fmt.Sprintf(`
	[
			{
				"type": "header",
				"text": {
					"type": "plain_text",
					"text": "%s\n"
				}
			},
			{
				"type": "context",
				"elements": [
					{
						"type": "mrkdwn",
						"text": "*%s*"
					}
				]
			},
			{
				"type": "divider"
			},
			{
				"type": "image",
				"image_url": "%s",
				"alt_text": ""
			},
			{
				"type": "divider"
			},
			{
				"type": "image",
				"image_url": "%s",
				"alt_text": ""
			},
			 {
				"type": "divider"
			}
	]

	`, reportTitle, dateRange, chartUrl, tableUrl)
	}
	return slackTemplate
}
func getSlackBlockForResult(title, result string) string {
	title = strings.Title(strings.ReplaceAll(title, "_", " "))
	title = removeDollarSymbolFromEventNames(title)
	slackBlock := fmt.Sprintf(`,
		{
			"type": "divider"
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*%s*\t\t*%s*"
			}
		}
		`, title, result)
	return slackBlock
}

func filterStringbyLastWord(displayCategory string, word string) string {
	arr := strings.Split(displayCategory, "_")
	// check last element of array is "metrics" if yes delete that
	if arr[len(arr)-1] == word || arr[len(arr)-1] == strings.Title(word) {
		arr = arr[:len(arr)-1]
	}
	return strings.Join(arr, "_")
}
func getEmojiForSlackByOperator(op string) string {
	if op == model.INCREASED_BY_MORE_THAN || op == model.PERCENTAGE_HAS_INCREASED_BY_MORE_THAN || op == model.IS_GREATER_THAN {
		return ":large_green_circle:"
	} else if op == model.DECREASED_BY_MORE_THAN || op == model.PERCENTAGE_HAS_DECREASED_BY_MORE_THAN || op == model.IS_LESS_THAN {
		return ":red_circle:"
	}
	return ""
}
func AddCommaToNumber(number string) string {
	if number == "" {
		return ""
	}
	arr := strings.Split(number, ".")
	output := arr[0]
	startOffset := 3
	for outputIndex := len(output); outputIndex > startOffset; {
		outputIndex -= 3
		output = output[:outputIndex] + "," + output[outputIndex:]
	}
	if len(arr) == 1 {
		return output
	}
	return output + "." + arr[1]
}
func removeDollarSymbolFromEventNames(title string) string {
	if len(title) > 0 {
		if title[0] == '$' {
			return title[1:]
		}
	}
	return title
}
func convertTimeFromSeconds(value float64) string {
	secondsInt := uint64(value)
	days := secondsInt / (24 * 3600)
	secondsInt = secondsInt % (24 * 3600)

	hours := secondsInt / 3600
	secondsInt = secondsInt % 3600

	minutes := secondsInt / 60
	secondsInt = secondsInt % 60

	seconds := secondsInt

	var daysStr, hoursStr, minstr, secstr string
	if days > 0 {
		daysStr = fmt.Sprintf("%d d ", days)
	}
	if hours > 0 {
		hoursStr = fmt.Sprintf("%d h ", hours)
	}
	if minutes > 0 {
		minstr = fmt.Sprintf("%d m ", minutes)
	}
	if seconds > 0 {
		secstr = fmt.Sprintf("%d s ", seconds)
	}

	return daysStr + hoursStr + minstr + secstr

}

// adding it here because import wasnt allowed (import cycle)
func GetDisplayEventNamesHandler(displayNames map[string]string) map[string]string {
	displayNameEvents := make(map[string]string)
	standardEvents := U.STANDARD_EVENTS_DISPLAY_NAMES
	for event, displayName := range standardEvents {
		displayNameEvents[event] = displayName
	}
	for event, displayName := range displayNames {
		displayNameEvents[event] = displayName
	}
	return displayNameEvents
}
