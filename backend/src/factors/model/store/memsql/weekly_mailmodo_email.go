package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const SECONDS_IN_A_WEEK int64 = 7 * 24 * 60 * 60
const METRIC_NOT_AVAILABLE = "NA"

type UserCount struct {
	UniqueUserCount int64 `json:"unique_user_count"`
}

func (store *MemSQL) GetUniqueUsersCountWithFormSubmissionEvent(projectId int64, startTimeStamp int64, endTimeStamp int64) (UserCount, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var count UserCount
	formSubmittedEventName, errCode := store.GetEventName(U.EVENT_NAME_FORM_SUBMITTED, projectId)
	if errCode == http.StatusNotFound {
		logCtx.Warn("Event name id for form submitted event not found")
		return count, nil
	} else if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for form submitted event")
		return count, fmt.Errorf("failed to get event name id for form submitted event")
	}

	db := C.GetServices().Db
	err := db.Table("events").Select("COUNT(DISTINCT(user_id)) as unique_user_count").
		Where("project_id=? and timestamp>=? and timestamp<=? and event_name_id =?", projectId, startTimeStamp, endTimeStamp, formSubmittedEventName.ID).
		Find(&count).Error
	if err != nil && err.Error() != "record not found" {
		logCtx.WithError(err).Error("Failed to get unique user count for form submission event")
		return count, err
	}

	return count, nil

}

func (store *MemSQL) GetUniqueUsersCountWithWebsiteSessionEvent(projectId int64, startTimeStamp int64, endTimeStamp int64) (UserCount, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var count UserCount
	sessionEventName, errCode := store.GetEventName(U.EVENT_NAME_SESSION, projectId)
	if errCode == http.StatusNotFound {
		logCtx.Warn("Event name id for website session event not found")
		return count, nil
	} else if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for website session event")
		return count, fmt.Errorf("failed to get event name id for website session event")
	}

	db := C.GetServices().Db
	err := db.Table("events").Select("COUNT(DISTINCT(user_id)) as unique_user_count").
		Where("project_id=? and timestamp>=? and timestamp<=? and event_name_id =?", projectId, startTimeStamp, endTimeStamp, sessionEventName.ID).
		Find(&count).Error
	if err != nil && err.Error() != "record not found" {
		logCtx.WithError(err).Error("Failed to get unique user count for website session event")
		return count, err
	}

	return count, nil

}

type DomainCount struct {
	UniqueDomainCount int64 `json:"unique_domain_count"`
}

func (store *MemSQL) GetUniqueFactorsDeanonIdentifiedCompaniesCount(projectId, startTimeStamp, endTimeStamp int64) (DomainCount, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	var count DomainCount
	err := db.Table("events").Select("COUNT(DISTINCT(JSON_EXTRACT_STRING(user_properties,'$6Signal_domain'))) as unique_domain_count").
		Where("project_id= ? and timestamp>=? and timestamp<=?", projectId, startTimeStamp, endTimeStamp).Find(&count).Error
	if err != nil && err.Error() != "record not found" {
		logCtx.WithError(err).Error("Failed to get unique domains count.")
		return count, err
	}

	return count, nil

}

func (store *MemSQL) GetTotalUniqueFactorsDeanonIdentifiedCompaniesCount(projectId int64) (DomainCount, error) {

	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	var count DomainCount
	err := db.Table("events").Select("COUNT(DISTINCT(JSON_EXTRACT_STRING(user_properties,'$6Signal_domain'))) as unique_domain_count").
		Where("project_id= ? ", projectId).Find(&count).Error
	if err != nil && err.Error() != "record not found" {
		logCtx.WithError(err).Error("Failed to get total unique domains count.")
		return count, err
	}

	return count, nil

}

type propCount struct {
	PropValue   string `json:"prop_value"`
	CountDomain int64  `json:"count_domain"`
}

func (store *MemSQL) GetFactorsDeanonPropertyTopValueCount(propName string, projectId, startTimeStamp, endTimeStamp int64) (propCount, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
		"propName":       propName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var propertyCount propCount

	sessionEventName, errCode := store.GetEventName(U.EVENT_NAME_SESSION, projectId)
	if errCode == http.StatusNotFound {
		logCtx.Warn("Event name id for session event not found")
		return propertyCount, nil
	} else if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for session event")
		return propertyCount, fmt.Errorf("failed to get event name id for session event")
	}

	db := C.GetServices().Db
	err := db.Limit(1).Table("events").
		Select("JSON_EXTRACT_STRING(user_properties, ?) as prop_value, COUNT(DISTINCT(JSON_EXTRACT_STRING(user_properties,'$6Signal_domain'))) as count_domain", propName).
		Where("project_id=? and timestamp>=? and timestamp<=? and event_name_id = ? and prop_value IS NOT NULL", projectId, startTimeStamp, endTimeStamp, sessionEventName.ID).
		Group("prop_value").Order("count_domain desc").Find(&propertyCount).Error
	if err != nil && err.Error() != "record not found" {
		logCtx.WithError(err).Error("Failed to get the property top value and count.")
		return propertyCount, err
	}

	return propertyCount, nil

}

type SessionTopPropertyValue struct {
	PropValue string `json:"prop_value"`
}

func (store *MemSQL) GetPropertyTopValueForSessionEvent(propName string, projectId, startTimeStamp, endTimeStamp int64) (SessionTopPropertyValue, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
		"propName":       propName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var propTopValue SessionTopPropertyValue

	sessionEventName, errCode := store.GetEventName(U.EVENT_NAME_SESSION, projectId)
	if errCode == http.StatusNotFound {
		logCtx.Warn("Event name id for session event not found")
		return propTopValue, nil
	} else if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for session event")
		return propTopValue, fmt.Errorf("failed to get event name id for session event")
	}

	db := C.GetServices().Db
	err := db.Limit(1).Table("events").Select("JSON_EXTRACT_STRING(properties, ?) as prop_value", propName).
		Where("project_id=? and timestamp>=? and timestamp<=? and event_name_id = ? and prop_value IS NOT NULL", projectId, startTimeStamp, endTimeStamp, sessionEventName.ID).
		Group("prop_value").Order("COUNT(*) desc").Find(&propTopValue).Error
	if err != nil && err.Error() != "record not found" {
		logCtx.WithError(err).Error("Failed to get property top value for session event")
		return propTopValue, err
	}

	return propTopValue, nil
}

func (store *MemSQL) GetMostActiveCompanies(projectId, startTimeStamp, endTimeStamp int64) ([]model.ActiveCompanyProp, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	activeCompaniesProp := make([]model.ActiveCompanyProp, 5)

	sessionEventName, errCode := store.GetEventName(U.EVENT_NAME_SESSION, projectId)
	if errCode == http.StatusNotFound {
		logCtx.Warn("Event name id for session event not found")
		return activeCompaniesProp, nil
	} else if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for session event")
		return activeCompaniesProp, fmt.Errorf("failed to get event name id for session event")
	}

	db := C.GetServices().Db

	err := db.Limit(5).Table("events").
		Select("JSON_EXTRACT_STRING(user_properties,'$6Signal_domain') as domain,JSON_EXTRACT_STRING(user_properties,'$6Signal_industry') as industry,JSON_EXTRACT_STRING(user_properties,'$6Signal_country') as country,JSON_EXTRACT_STRING(user_properties,'$6Signal_revenue_range')  as revenue").
		Where("project_id= ? and timestamp>= ? and timestamp<= ? and event_name_id=? and domain IS NOT NULL and domain !='' ", projectId, startTimeStamp, endTimeStamp, sessionEventName.ID).
		Group("domain").Order("COUNT(*) desc").Find(&activeCompaniesProp).Error
	if err != nil && err.Error() != "record not found" {
		logCtx.WithError(err).Error("Failed to get most active companies")
		return activeCompaniesProp, err
	}

	return activeCompaniesProp, nil

}

func (store *MemSQL) GetWeeklyMailmodoEmailsMetrics(projectId, startTimeStamp, endTimeStamp int64) (model.WeeklyMailmodoEmailMetrics, error) {

	var metrics model.WeeklyMailmodoEmailMetrics

	lastWeekStartTimeStamp := startTimeStamp - SECONDS_IN_A_WEEK
	lastWeekEndTimeStamp := endTimeStamp - SECONDS_IN_A_WEEK

	sessionUniqueUserCount, err := store.GetUniqueUsersCountWithWebsiteSessionEvent(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	formSubmittedUniqueUserCount, err := store.GetUniqueUsersCountWithFormSubmissionEvent(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	factorsDeanonIdentifiedCompaniesCount, err := store.GetUniqueFactorsDeanonIdentifiedCompaniesCount(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	factorsDeanonIdentifiedCompaniesCountLastWeek, err := store.GetUniqueFactorsDeanonIdentifiedCompaniesCount(projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	totalFactorsDeanonIdentifiedCompaniesCount, err := store.GetTotalUniqueFactorsDeanonIdentifiedCompaniesCount(projectId)
	if err != nil {
		return metrics, err
	}

	industryTopValue, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_INDUSTRY, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	industryTopValueLastWeek, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_INDUSTRY, projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	employeeRangeTopValue, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_EMPLOYEE_RANGE, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	employeeRangeTopValueLastWeek, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_EMPLOYEE_RANGE, projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	countryTopValue, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_COUNTRY, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	countryTopValueLastWeek, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_COUNTRY, projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	revenueRangeTopValue, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_REVENUE_RANGE, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	revenueRangeTopValueLastWeek, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_REVENUE_RANGE, projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	sessionCount, err := store.GetSessionCount(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	channelTopValue, err := store.GetPropertyTopValueForSessionEvent(U.EP_CHANNEL, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	campaignTopValue, err := store.GetPropertyTopValueForSessionEvent(U.EP_CAMPAIGN, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	sourceTopValue, err := store.GetPropertyTopValueForSessionEvent(U.EP_SOURCE, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	activeCompaniesProp, err := store.GetMostActiveCompanies(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	metrics.SessionUniqueUserCount = sessionUniqueUserCount.UniqueUserCount
	metrics.FormSubmittedUniqueUserCount = formSubmittedUniqueUserCount.UniqueUserCount
	metrics.IdentifiedCompaniesCount = factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
	metrics.TotalIdentifiedCompaniesCount = totalFactorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
	metrics.SessionCount = sessionCount.SessionCount
	metrics.TopChannel = GetValidatedMetricForStringValue(channelTopValue.PropValue)
	metrics.TopSource = GetValidatedMetricForStringValue(sourceTopValue.PropValue)
	metrics.TopCampaign = GetValidatedMetricForStringValue(campaignTopValue.PropValue)
	metrics.Industry.TopValue = GetValidatedMetricForStringValue(industryTopValue.PropValue)
	metrics.Industry.TopValueLastWeek = GetValidatedMetricForStringValue(industryTopValueLastWeek.PropValue)
	metrics.Country.TopValue = GetValidatedMetricForStringValue(countryTopValue.PropValue)
	metrics.Country.TopValueLastWeek = GetValidatedMetricForStringValue(countryTopValueLastWeek.PropValue)
	metrics.EmployeeRange.TopValue = GetValidatedMetricForStringValue(employeeRangeTopValue.PropValue)
	metrics.EmployeeRange.TopValueLastWeek = GetValidatedMetricForStringValue(employeeRangeTopValueLastWeek.PropValue)
	metrics.RevenueRange.TopValue = GetValidatedMetricForStringValue(revenueRangeTopValue.PropValue)
	metrics.RevenueRange.TopValueLastWeek = GetValidatedMetricForStringValue(revenueRangeTopValueLastWeek.PropValue)

	for i, v := range activeCompaniesProp {
		switch i {
		case 0:
			metrics.Company1.Domain = GetValidatedMetricForStringValue(v.Domain)
			metrics.Company1.Country = GetValidatedMetricForStringValue(v.Country)
			metrics.Company1.Industry = GetValidatedMetricForStringValue(v.Industry)
			metrics.Company1.Revenue = GetValidatedMetricForStringValue(v.Revenue)
		case 1:
			metrics.Company2.Domain = GetValidatedMetricForStringValue(v.Domain)
			metrics.Company2.Country = GetValidatedMetricForStringValue(v.Country)
			metrics.Company2.Industry = GetValidatedMetricForStringValue(v.Industry)
			metrics.Company2.Revenue = GetValidatedMetricForStringValue(v.Revenue)
		case 2:
			metrics.Company3.Domain = GetValidatedMetricForStringValue(v.Domain)
			metrics.Company3.Country = GetValidatedMetricForStringValue(v.Country)
			metrics.Company3.Industry = GetValidatedMetricForStringValue(v.Industry)
			metrics.Company3.Revenue = GetValidatedMetricForStringValue(v.Revenue)
		case 3:
			metrics.Company4.Domain = GetValidatedMetricForStringValue(v.Domain)
			metrics.Company4.Country = GetValidatedMetricForStringValue(v.Country)
			metrics.Company4.Industry = GetValidatedMetricForStringValue(v.Industry)
			metrics.Company4.Revenue = GetValidatedMetricForStringValue(v.Revenue)
		case 4:
			metrics.Company5.Domain = GetValidatedMetricForStringValue(v.Domain)
			metrics.Company5.Country = GetValidatedMetricForStringValue(v.Country)
			metrics.Company5.Industry = GetValidatedMetricForStringValue(v.Industry)
			metrics.Company5.Revenue = GetValidatedMetricForStringValue(v.Revenue)

		}
	}

	if factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount <= 0 && factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount <= 0 {

		metrics.Industry.TopValuePercent = 0
		metrics.Industry.TopValuePercentLastWeek = 0
		metrics.Country.TopValuePercent = 0
		metrics.Country.TopValuePercentLastWeek = 0
		metrics.EmployeeRange.TopValuePercent = 0
		metrics.EmployeeRange.TopValuePercentLastWeek = 0
		metrics.RevenueRange.TopValuePercent = 0
		metrics.RevenueRange.TopValuePercentLastWeek = 0

	} else if factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount > 0 && factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount <= 0 {

		metrics.Industry.TopValuePercent = (industryTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
		metrics.Industry.TopValuePercentLastWeek = 0
		metrics.Country.TopValuePercent = (countryTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
		metrics.Country.TopValuePercentLastWeek = 0
		metrics.EmployeeRange.TopValuePercent = (employeeRangeTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
		metrics.EmployeeRange.TopValuePercentLastWeek = 0
		metrics.RevenueRange.TopValuePercent = (revenueRangeTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
		metrics.RevenueRange.TopValuePercentLastWeek = 0

	} else if factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount <= 0 && factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount > 0 {

		metrics.Industry.TopValuePercent = 0
		metrics.Industry.TopValuePercentLastWeek = (industryTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount
		metrics.Country.TopValuePercent = 0
		metrics.Country.TopValuePercentLastWeek = (countryTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount
		metrics.EmployeeRange.TopValuePercent = 0
		metrics.EmployeeRange.TopValuePercentLastWeek = (employeeRangeTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount
		metrics.RevenueRange.TopValuePercent = 0
		metrics.RevenueRange.TopValuePercentLastWeek = (revenueRangeTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount

	} else {
		metrics.Industry.TopValuePercent = (industryTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
		metrics.Industry.TopValuePercentLastWeek = (industryTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount
		metrics.Country.TopValuePercent = (countryTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
		metrics.Country.TopValuePercentLastWeek = (countryTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount
		metrics.EmployeeRange.TopValuePercent = (employeeRangeTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
		metrics.EmployeeRange.TopValuePercentLastWeek = (employeeRangeTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount
		metrics.RevenueRange.TopValuePercent = (revenueRangeTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount
		metrics.RevenueRange.TopValuePercentLastWeek = (revenueRangeTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount
	}

	return metrics, nil

}

func GetValidatedMetricForStringValue(metric string) string {
	if metric == "" {
		return METRIC_NOT_AVAILABLE
	}

	return metric
}
