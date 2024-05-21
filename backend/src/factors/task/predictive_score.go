package task

import (
	"bufio"
	"encoding/json"
	"factors/filestore"
	"factors/merge"
	M "factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	"factors/pull"
	U "factors/util"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

var props = map[string]string{
	"ep#$initial_referrer_domain":                                "first",
	"ep#$is_first_session":                                       "first",
	"ep#$page_count":                                             "total",
	"ep#$session_spent_time":                                     "total",
	"ep#$initial_page_load_time":                                 "average",
	"ep#$initial_page_spent_time":                                "average",
	"ep#$initial_page_scroll_percent":                            "average",
	"ep#$page_load_time":                                         "average",
	"ep#$page_spent_time":                                        "average",
	"ep#$page_scroll_percent":                                    "average",
	"up#$initial_page_load_time":                                 "average",
	"up#$initial_page_spent_time":                                "average",
	"up#$initial_page_scroll_percent":                            "average",
	"up#$latest_page_load_time":                                  "average",
	"up#$latest_page_spent_time":                                 "average",
	"up#$latest_page_scroll_percent":                             "average",
	"up#$continent":                                              "last",
	"up#$country":                                                "last",
	"up#$city":                                                   "last",
	"up#$source":                                                 "last",
	"up#$campaign":                                               "last",
	"up#$medium":                                                 "last",
	"up#$content":                                                "last",
	"up#$initial_referrer_domain":                                "last",
	"up#$latest_referrer_domain":                                 "last",
	"up#$latest_channel":                                         "last",
	"up#$initial_channel":                                        "last",
	"up#$latest_source":                                          "last",
	"up#$latest_medium":                                          "last",
	"up#$latest_content":                                         "last",
	"up#$latest_campaign":                                        "last",
	"up#$initial_source":                                         "last",
	"up#$initial_medium":                                         "last",
	"up#$initial_content":                                        "last",
	"up#$initial_campaign":                                       "last",
	"up#$latest_adgroup_id":                                      "last",
	"up#$latest_campaign_id":                                     "last",
	"up#$latest_term":                                            "last",
	"up#$initial_adgroup_id":                                     "last",
	"up#$initial_campaign_id":                                    "last",
	"up#$initial_term":                                           "last",
	"up#$browser":                                                "last",
	"up#$browser_version":                                        "last",
	"up#$device_name":                                            "last",
	"up#$6Signal_country":                                        "last",
	"up#$6Signal_country_iso_code":                               "last",
	"up#$6Signal_state":                                          "last",
	"up#$6Signal_city":                                           "last",
	"up#$6Signal_region":                                         "last",
	"up#$6Signal_employee_range":                                 "last",
	"up#$6Signal_industry":                                       "last",
	"up#$6Signal_revenue_range":                                  "last",
	"up#$6Signal_employee_count":                                 "last",
	"up#$6Signal_annual_revenue":                                 "last",
	"up#$li_headquarter":                                         "last",
	"up#$li_total_ad_view_count":                                 "last",
	"up#$li_total_ad_click_count":                                "last",
	"up#$hubspot_contact_lifecyclestage":                         "last",
	"up#$hubspot_contact_rh_meeting_status":                      "last",
	"up#$hubspot_contact_hs_timezone":                            "last",
	"up#$hubspot_contact_jobtitle":                               "last",
	"up#$hubspot_contact_rh_meeting_type":                        "last",
	"up#$hubspot_contact_rh_no_show":                             "last",
	"up#$hubspot_contact_company_annual_revenue":                 "last",
	"up#$hubspot_contact_company":                                "last",
	"up#$hubspot_contact_state":                                  "last",
	"up#$hubspot_contact_hs_analytics_num_page_views":            "last",
	"up#$hubspot_contact_hs_analytics_num_visits":                "last",
	"up#$hubspot_contact_days_to_close":                          "last",
	"up#$hubspot_contact_total_revenue":                          "last",
	"up#$hubspot_contact_job_function":                           "last",
	"up#$hubspot_contact_icp":                                    "last",
	"up#$hubspot_contact_icp_industry_category":                  "last",
	"up#$hubspot_contact_employee_range":                         "last",
	"up#$hubspot_contact_annualrevenue":                          "last",
	"up#$hubspot_contact_industry":                               "last",
	"up#$hubspot_contact_demo_booked_date":                       "last",
	"up#$hubspot_contact_hs_email_delivered":                     "last",
	"up#$hubspot_company_type":                                   "last",
	"up#$hubspot_company_first_conversion_event_name":            "last",
	"up#$hubspot_company_web_technologies":                       "last",
	"up#$hubspot_company_churned":                                "last",
	"up#$hubspot_company_country":                                "last",
	"up#$hubspot_company_hs_annual_revenue_currency_code":        "last",
	"up#$hubspot_company_founded_year":                           "last",
	"up#$hubspot_company_state":                                  "last",
	"up#$hubspot_company_city":                                   "last",
	"up#$hubspot_company_timezone":                               "last",
	"up#$hubspot_company_hs_analytics_latest_source":             "last",
	"up#$hubspot_company_industry":                               "last",
	"up#$hubspot_company_total_money_raised":                     "last",
	"up#$hubspot_company_lifecyclestage":                         "last",
	"up#$hubspot_company_is_public":                              "last",
	"up#$hubspot_company_hs_pipeline":                            "last",
	"up#$hubspot_company_hs_num_contacts_with_buying_roles":      "last",
	"up#$hubspot_company_annualrevenue":                          "last",
	"up#$hubspot_company_hs_analytics_num_page_views":            "last",
	"up#$hubspot_company_hs_num_decision_makers":                 "last",
	"up#$hubspot_company_num_conversion_events":                  "last",
	"up#$hubspot_company_hs_analytics_num_visits":                "last",
	"up#$hubspot_company_total_revenue":                          "last",
	"up#$hubspot_company_arpu":                                   "last",
	"up#$hubspot_company_numberofemployees":                      "last",
	"up#$hubspot_company_hs_total_deal_value":                    "last",
	"up#$hubspot_company_outbound_company":                       "last",
	"up#$hubspot_company_intent":                                 "last",
	"up#$hubspot_company_icp":                                    "last",
	"up#$hubspot_company_inbound_outbound":                       "last",
	"up#$hubspot_company_meeting_status":                         "last",
	"up#$hubspot_company_engagements_last_meeting_booked_medium": "last",
	"up#$hubspot_company_icp_industry_category":                  "last",
	"up#$hubspot_company_hs_lead_status":                         "last",
	"up#$hubspot_company_revenue_growth":                         "last",
	"up#$hubspot_company_twitterfollowers":                       "last",
	"up#$hubspot_company_hs_last_booked_meeting_date":            "last",
	"up#$hubspot_deal_dealstage":                                 "last",
	"up#$hubspot_deal_hs_priority":                               "last",
	"up#$hubspot_deal_product":                                   "last",
	"up#$hubspot_deal_hs_is_closed_won":                          "last",
	"up#$hubspot_deal_hs_analytics_latest_source":                "last",
	"up#$hubspot_deal_hs_analytics_source":                       "last",
	"up#$hubspot_deal_pipeline":                                  "last",
	"up#$hubspot_deal_hs_is_closed":                              "last",
	"up#$hubspot_deal_amount_in_home_currency":                   "last",
	"up#$hubspot_deal_num_associated_contacts":                   "last",
	"up#$hubspot_deal_days_to_close":                             "last",
	"up#$hubspot_deal_hs_deal_stage_probability":                 "last",
	"up#$hubspot_deal_num_contacted_notes":                       "last",
	"up#$hubspot_deal_hs_closed_amount_in_home_currency":         "last",
	"up#$hubspot_deal_amount":                                    "last",
	"up#$hubspot_deal_contract_duration":                         "last",
	"up#$hubspot_deal_renewal_frequency":                         "last",
	"up#$hubspot_deal_hs_next_step":                              "last",
	"up#$hubspot_deal_inbound_outbound":                          "last",
	"up#$hubspot_deal_outbound_source":                           "last",
	"up#$hubspot_deal_hs_campaign":                               "last",
	"up#$hubspot_deal_hs_arr":                                    "last",
	"up#$hubspot_deal_hs_tcv":                                    "last",
	"up#$hubspot_deal_hs_forecast_probability":                   "last",
	"up#$hubspot_deal_hs_mrr":                                    "last",
	"up#$hubspot_deal_hs_acv":                                    "last"}

var minTimestampCol string = "minEventTimestamp"
var maxTimestampCol string = "maxEventTimestamp"
var IdColumnName string = "AccountId"

func PredictiveScoring2(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	logCtx := log.WithField("projectId", projectId)

	status := make(map[string]interface{})

	tmpCloudManager := configs["tmpCloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)
	lookbackWindow := configs["lookback"].(int)
	startTimestamp := *(configs["startTimestamp"].(*int64))
	endTimestamp := *(configs["endTimestamp"].(*int64))

	// get id number for aaccount ids
	domain_group, httpStatus := store.GetStore().GetGroup(projectId, M.GROUP_NAME_DOMAINS)
	if httpStatus != http.StatusFound {
		err := fmt.Errorf("failed to get existing groups (%s) for project (%d)", M.GROUP_NAME_DOMAINS, projectId)
		logCtx.WithField("err_code", status).Error(err)
		status["err"] = err.Error()
		return status, false
	}
	idNum := domain_group.ID

	startDayTimestamp := startTimestamp
	endDayTimestamp := endTimestamp

	accountInfos := make(map[string]map[string]interface{})
	userPropCounts := make(map[string]map[string]int)

	timestamp := startDayTimestamp
	for timestamp <= endDayTimestamp {
		partFilesDir, _ := pull.GetDailyArchiveFilePathAndName(archiveCloudManager, U.DataTypeEvent, U.EVENTS_FILENAME_PREFIX, projectId, timestamp, 0, 0)
		listFiles := (*archiveCloudManager).ListFiles(partFilesDir)
		for _, partFileFullName := range listFiles {
			partFNamelist := strings.Split(partFileFullName, "/")
			partFileName := partFNamelist[len(partFNamelist)-1]
			if !strings.HasPrefix(partFileName, U.EVENTS_FILENAME_PREFIX) {
				continue
			}

			log.Infof("Reading daily file :%s, %s", partFilesDir, partFileName)
			file, err := (*archiveCloudManager).Get(partFilesDir, partFileName)
			if err != nil {
				log.Error(err)
				status["err"] = err.Error()
				return status, false
			}

			scanner := bufio.NewScanner(file)
			const maxCapacity = 30 * 1024 * 1024
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			for scanner.Scan() {
				line := scanner.Text()

				var eventDetails *P.CounterEventFormat
				err = json.Unmarshal([]byte(line), &eventDetails)
				if err != nil {
					log.Error(err)
					status["err"] = err.Error()
					return status, false
				}
				AccID := merge.GetAptId(eventDetails, idNum)
				if _, ok := accountInfos[AccID]; !ok {
					newMap := make(map[string]interface{})
					newMap[IdColumnName] = AccID
					newMap[minTimestampCol] = eventDetails.EventTimestamp
					newMap[maxTimestampCol] = eventDetails.EventTimestamp
					accountInfos[AccID] = newMap
					userPropCounts[AccID] = make(map[string]int)
				}
				dataPoint := accountInfos[AccID]

				eName := "ev#" + eventDetails.EventName
				if freq, ok := dataPoint[eName]; !ok {
					dataPoint[eName] = 1
				} else {
					dataPoint[eName] = freq.(int64) + 1
				}

				evType := "middle"
				if eventDetails.EventTimestamp < dataPoint[minTimestampCol].(int64) {
					dataPoint[minTimestampCol] = eventDetails.EventTimestamp
					evType = "first"
				} else if eventDetails.EventTimestamp > dataPoint[maxTimestampCol].(int64) {
					dataPoint[maxTimestampCol] = eventDetails.EventTimestamp
					evType = "last"
				}

				propCounts := userPropCounts[AccID]
				for uKey, uVal := range eventDetails.UserProperties {
					uKey = "up#" + uKey
					if uVal == "" || uVal == nil {
						continue
					}
					if val, ok := props[uKey]; !ok {
						continue
					} else {
						propCounts[uKey] += 1
						if val == "pass" {
							continue
						} else if val == "total" || val == "average" {
							floatVal, err := U.GetPropertyValueAsFloat64(uVal)
							if err != nil {
								log.Error("failed getting interface float value")
								status["err"] = err.Error()
								return status, false
							}
							if _, ok := dataPoint[uKey]; !ok {
								dataPoint[uKey] = 0.0
							}
							dataPoint[uKey] = dataPoint[uKey].(float64) + floatVal
						} else if val == "first" || val == "last" {
							if evType != val {
								continue
							}
							dataPoint[uKey] = uVal
						}
					}
				}
				for eKey, eVal := range eventDetails.EventProperties {
					eKey = "ep#" + eKey
					if eVal == "" || eVal == nil {
						continue
					}
					if val, ok := props[eKey]; !ok {
						continue
					} else {
						propCounts[eKey] += 1
						if _, ok := dataPoint[eKey]; !ok {
							dataPoint[eKey] = eVal
							continue
						}
						if val == "pass" {
							continue
						} else if val == "total" || val == "average" {
							floatVal, err := U.GetPropertyValueAsFloat64(eVal)
							if err != nil {
								log.Error("failed getting interface float value")
								status["err"] = err.Error()
								return status, false
							}
							if _, ok := dataPoint[eKey]; !ok {
								dataPoint[eKey] = 0.0
							}
							dataPoint[eKey] = dataPoint[eKey].(float64) + floatVal
						} else if val == "first" || val == "last" {
							if evType != val {
								continue
							}
							dataPoint[eKey] = eVal
						}
					}
				}
			}
		}
		timestamp += U.Per_day_epoch
	}

	newStartDayTimestamp := startDayTimestamp - (int64(lookbackWindow) * U.Per_day_epoch)
	newEndDayTimestamp := startDayTimestamp - 1
	timestamp = newStartDayTimestamp
	// var accountInfos = make(map[string]map[string]interface{})
	for timestamp <= newEndDayTimestamp {
		partFilesDir, _ := pull.GetDailyArchiveFilePathAndName(archiveCloudManager, U.DataTypeEvent, U.EVENTS_FILENAME_PREFIX, projectId, timestamp, 0, 0)
		listFiles := (*archiveCloudManager).ListFiles(partFilesDir)
		for _, partFileFullName := range listFiles {
			partFNamelist := strings.Split(partFileFullName, "/")
			partFileName := partFNamelist[len(partFNamelist)-1]
			if !strings.HasPrefix(partFileName, U.EVENTS_FILENAME_PREFIX) {
				continue
			}

			log.Infof("Reading daily file :%s, %s", partFilesDir, partFileName)
			file, err := (*archiveCloudManager).Get(partFilesDir, partFileName)
			if err != nil {
				log.Error(err)
				status["err"] = err.Error()
				return status, false
			}

			scanner := bufio.NewScanner(file)
			const maxCapacity = 30 * 1024 * 1024
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			for scanner.Scan() {
				line := scanner.Text()

				var eventDetails *P.CounterEventFormat
				err = json.Unmarshal([]byte(line), &eventDetails)
				if err != nil {
					log.Error(err)
					status["err"] = err.Error()
					return status, false
				}
				AccID := merge.GetAptId(eventDetails, idNum)

				if _, ok := accountInfos[AccID]; !ok {
					continue
				}
				dataPoint := accountInfos[AccID]

				eName := "ev#" + eventDetails.EventName
				if freq, ok := dataPoint[eName]; !ok {
					dataPoint[eName] = 1
				} else {
					dataPoint[eName] = freq.(int64) + 1
				}

				evType := "middle"
				if eventDetails.EventTimestamp < dataPoint[minTimestampCol].(int64) {
					dataPoint[minTimestampCol] = eventDetails.EventTimestamp
					evType = "first"
				} else if eventDetails.EventTimestamp > dataPoint[maxTimestampCol].(int64) {
					dataPoint[maxTimestampCol] = eventDetails.EventTimestamp
					evType = "last"
				}

				propCounts := userPropCounts[AccID]
				for uKey, uVal := range eventDetails.UserProperties {
					uKey = "up#" + uKey
					if uVal == "" || uVal == nil {
						continue
					}
					if val, ok := props[uKey]; !ok {
						continue
					} else {
						propCounts[uKey] += 1
						if val == "pass" {
							continue
						} else if val == "total" || val == "average" {
							floatVal, err := U.GetPropertyValueAsFloat64(uVal)
							if err != nil {
								log.Error("failed getting interface float value")
								status["err"] = err.Error()
								return status, false
							}
							if _, ok := dataPoint[uKey]; !ok {
								dataPoint[uKey] = 0.0
							}
							dataPoint[uKey] = dataPoint[uKey].(float64) + floatVal
						} else if val == "first" || val == "last" {
							if evType != val {
								continue
							}
							dataPoint[uKey] = uVal
						}
					}
				}
				for eKey, eVal := range eventDetails.EventProperties {
					eKey = "ep#" + eKey
					if eVal == "" || eVal == nil {
						continue
					}
					if val, ok := props[eKey]; !ok {
						continue
					} else {
						propCounts[eKey] += 1
						if val == "pass" {
							continue
						} else if val == "total" || val == "average" {
							floatVal, err := U.GetPropertyValueAsFloat64(eVal)
							if err != nil {
								log.Error("failed getting interface float value")
								status["err"] = err.Error()
								return status, false
							}
							if _, ok := dataPoint[eKey]; !ok {
								dataPoint[eKey] = 0.0
							}
							dataPoint[eKey] = dataPoint[eKey].(float64) + floatVal
						} else if val == "first" || val == "last" {
							if evType != val {
								continue
							}
							dataPoint[eKey] = eVal
						}
					}
				}
			}
		}
		timestamp += U.Per_day_epoch
	}

	fileDir := (*tmpCloudManager).GetProjectDir(projectId)
	fileDir = fileDir + "pred_score_rfc/"
	fileName := "training_data_golang.txt"
	cloudWriter, err := (*tmpCloudManager).GetWriter(fileDir, fileName)
	if err != nil {
		log.WithFields(log.Fields{"fileDir": fileDir, "fileName": fileName}).Error("unable to get writer for file")
		status["err"] = err.Error()
		return status, false
	}

	countIds := 0
	for _, dataPoint := range accountInfos {
		if dataPointBytes, err := json.Marshal(dataPoint); err != nil {
			log.WithFields(log.Fields{"dataPoint": dataPoint, "err": err}).Error("Marshal failed")
			status["err"] = err.Error()
			return status, false
		} else {
			lineWrite := string(dataPointBytes)
			if _, err := io.WriteString(cloudWriter, lineWrite+"\n"); err != nil {
				mineLog.WithFields(log.Fields{"line": lineWrite, "err": err}).Error("Unable to write to file.")
				status["err"] = err.Error()
				return status, false
			}
		}
		countIds += 1
	}
	err = cloudWriter.Close()
	if err != nil {
		log.WithFields(log.Fields{"fileDir": fileDir, "fileName": fileName}).Error("unable to close writer for file")
		status["err"] = err.Error()
		return status, false
	}

	return status, true
}
