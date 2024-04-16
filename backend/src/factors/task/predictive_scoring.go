package task

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	"factors/merge"
	M "factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"io"
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"
)

const NO_OF_UNIQUE_EVENTS_AND_PROPERTY_VALUES = 500
const maxCapacity = 30 * 1024 * 1024
const idColumnName = "AccountId"
const eventsColumnName = "EventName"
const timestampColumnName = "EventTimestamp"

func PredictiveScoring(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	logCtx := log.WithField("projectId", projectId)

	status := make(map[string]interface{})

	beamConfig := configs["beamConfig"].(*merge.RunBeamConfig)
	tmpCloudManager := configs["tmpCloudManager"].(*filestore.FileManager)
	sortedCloudManager := configs["sortedCloudManager"].(*filestore.FileManager)
	modelCloudManager := configs["modelCloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	lookbackWindow := configs["lookback"].(int)
	startTimestamp := *(configs["startTimestamp"].(*int64))
	endTimestamp := *(configs["endTimestamp"].(*int64))
	createNewProps := configs["createNewProps"].(bool)

	propsToFilter := []string{idColumnName, eventsColumnName, timestampColumnName, "ep#$initial_referrer_domain", "ep#$is_first_session",
		"ep#$source", "ep#$medium", "ep#$campaign", "ep#$campaign_id", "ep#$term", "ep#$adgroup_id", "ep#$content", "ep#$creative",
		"ep#$keyword", "ep#$keyword_match_type", "ep#$session_count", "ep#$page_count", "ep#$session_spent_time",
		"ep#$initial_page_load_time", "ep#$initial_page_spent_time", "ep#$initial_page_scroll_percent", "ep#$is_page_view",
		"ep#$page_load_time", "ep#$page_spent_time", "ep#$page_scroll_percent", "up#$initial_page_url", "up#$initial_page_load_time",
		"up#$initial_page_spent_time", "up#$initial_page_scroll_percent", "up#$latest_page_url", "up#$latest_page_load_time",
		"up#$latest_page_spent_time", "up#$latest_page_scroll_percent", "up#$continent", "up#$country", "up#$city", "up#$source",
		"up#$campaign", "up#$medium", "up#$content", "up#$initial_referrer_domain", "up#$latest_referrer_domain",
		"up#$latest_channel", "up#$initial_channel", "up#$latest_source", "up#$latest_medium", "up#$latest_content",
		"up#$latest_campaign", "up#$initial_source", "up#$initial_medium", "up#$initial_content", "up#$initial_campaign",
		"up#$latest_adgroup_id", "up#$latest_campaign_id", "up#$latest_term", "up#$initial_adgroup_id", "up#$initial_campaign_id",
		"up#$initial_term", "up#$browser", "up#$browser_version", "up#$device_name", "up#$6Signal_country",
		"up#$6Signal_country_iso_code", "up#$6Signal_state", "up#$6Signal_city", "up#$6Signal_region", "up#$6Signal_employee_range",
		"up#$6Signal_industry", "up#$6Signal_revenue_range", "up#$6Signal_employee_count", "up#$6Signal_annual_revenue",
		"up#$li_headquarter", "up#$li_total_ad_view_count", "up#$li_total_ad_click_count", "ep#$g2_tag", "ep#$g2_product_ids",
		"ep#$g2_visitor_country", "ep#$g2_category_ids", "ep#$g2_visitor_state", "ep#$g2_visitor_city", "up#$hubspot_contact_lifecyclestage",
		"up#$hubspot_contact_rh_meeting_status", "up#$hubspot_contact_hs_timezone", "up#$hubspot_contact_jobtitle",
		"up#$hubspot_contact_rh_meeting_type", "up#$hubspot_contact_rh_no_show", "up#$hubspot_contact_company_annual_revenue",
		"up#$hubspot_contact_company", "up#$hubspot_contact_state", "up#$hubspot_contact_hs_analytics_num_page_views",
		"up#$hubspot_contact_hs_analytics_num_visits", "up#$hubspot_contact_days_to_close", "up#$hubspot_contact_total_revenue",
		"up#$hubspot_contact_job_function", "up#$hubspot_contact_icp", "up#$hubspot_contact_icp_industry_category",
		"up#$hubspot_contact_employee_range", "up#$hubspot_contact_annualrevenue", "up#$hubspot_contact_industry",
		"up#$hubspot_contact_demo_booked_date", "up#$hubspot_contact_hs_email_delivered", "up#$hubspot_company_type",
		"up#$hubspot_company_first_conversion_event_name", "up#$hubspot_company_web_technologies", "up#$hubspot_company_churned",
		"up#$hubspot_company_country", "up#$hubspot_company_hs_annual_revenue_currency_code", "up#$hubspot_company_founded_year",
		"up#$hubspot_company_state", "up#$hubspot_company_city", "up#$hubspot_company_timezone", "up#$hubspot_company_hs_analytics_latest_source",
		"up#$hubspot_company_industry", "up#$hubspot_company_total_money_raised", "up#$hubspot_company_lifecyclestage",
		"up#$hubspot_company_is_public", "up#$hubspot_company_hs_pipeline", "up#$hubspot_company_hs_num_contacts_with_buying_roles",
		"up#$hubspot_company_annualrevenue", "up#$hubspot_company_hs_analytics_num_page_views", "up#$hubspot_company_hs_num_decision_makers", "up#$hubspot_company_num_conversion_events", "up#$hubspot_company_hs_analytics_num_visits", "up#$hubspot_company_total_revenue", "up#$hubspot_company_arpu", "up#$hubspot_company_numberofemployees", "up#$hubspot_company_hs_total_deal_value", "up#$hubspot_company_outbound_company", "up#$hubspot_company_intent", "up#$hubspot_company_icp", "up#$hubspot_company_inbound_outbound", "up#$hubspot_company_meeting_status", "up#$hubspot_company_engagements_last_meeting_booked_medium", "up#$hubspot_company_icp_industry_category", "up#$hubspot_company_hs_lead_status", "up#$hubspot_company_revenue_growth", "up#$hubspot_company_twitterfollowers", "up#$hubspot_company_hs_last_booked_meeting_date", "up#$hubspot_deal_dealstage", "up#$hubspot_deal_hs_priority", "up#$hubspot_deal_product", "up#$hubspot_deal_hs_is_closed_won", "up#$hubspot_deal_hs_analytics_latest_source", "up#$hubspot_deal_hs_analytics_source", "up#$hubspot_deal_pipeline", "up#$hubspot_deal_hs_is_closed", "up#$hubspot_deal_amount_in_home_currency", "up#$hubspot_deal_num_associated_contacts", "up#$hubspot_deal_days_to_close", "up#$hubspot_deal_hs_deal_stage_probability", "up#$hubspot_deal_num_contacted_notes", "up#$hubspot_deal_hs_closed_amount_in_home_currency", "up#$hubspot_deal_amount", "up#$hubspot_deal_contract_duration", "up#$hubspot_deal_renewal_frequency", "up#$hubspot_deal_hs_next_step", "up#$hubspot_deal_inbound_outbound", "up#$hubspot_deal_outbound_source", "up#$hubspot_deal_hs_campaign", "up#$hubspot_deal_hs_arr", "up#$hubspot_deal_hs_tcv", "up#$hubspot_deal_hs_forecast_probability", "up#$hubspot_deal_hs_mrr", "up#$hubspot_deal_hs_acv"}

	mapProp := make(map[string]bool)
	for _, prop := range propsToFilter {
		mapProp[prop] = true
	}

	// get id number for aaccount ids
	domain_group, httpStatus := store.GetStore().GetGroup(projectId, M.GROUP_NAME_DOMAINS)
	if httpStatus != http.StatusFound {
		err := fmt.Errorf("failed to get existing groups (%s) for project (%d)", M.GROUP_NAME_DOMAINS, projectId)
		logCtx.WithField("err_code", status).Error(err)
		status["err"] = err.Error()
		return status, false
	}
	idNum := domain_group.ID

	// get events file for the week
	efCloudPath, efCloudName, err := merge.MergeAndWriteSortedFile(projectId, U.DataTypeEvent, "", startTimestamp, endTimestamp,
		archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, false, domain_group.ID, false, true, true, nil)
	if err != nil {
		status["err"] = err.Error()
		logCtx.WithError(err).Error("Failed creating events file")
		return status, false
	}

	// get all active accounts for the week
	accountIds, err := getIdsFromEventsFile(efCloudPath, efCloudName, sortedCloudManager, idNum)
	if err != nil {
		logCtx.WithError(err).Error("Failed getting account ids from events file")
		status["err"] = err.Error()
		return status, false
	}

	// get past 90 days timeline + the week for all accounts (active in that week)
	newStart := startTimestamp - (int64(lookbackWindow) * U.Per_day_epoch)
	dataFilePath, dataFileName, err := merge.MergeAndWriteSortedFile(projectId, U.DataTypeEvent, "", newStart, endTimestamp,
		archiveCloudManager, tmpCloudManager, tmpCloudManager, diskManager, beamConfig, false, domain_group.ID, false, true, true, accountIds)
	if err != nil {
		status["err"] = err.Error()
		logCtx.WithError(err).Error("Failed creating events file")
		return status, false
	}

	// convert timestamps to UTC
	projectDetails, _ := store.GetStore().GetProject(projectId)
	if projectDetails.TimeZone != "" {
		log.Infof("Project Timezone not UTC - Converting timestamps to UTC from project timezone(%s)", projectDetails.TimeZone)
		offset := U.FindOffsetInUTC(U.TimeZoneString(projectDetails.TimeZone))
		startTimestamp = startTimestamp + int64(offset)
		endTimestamp = endTimestamp + int64(offset)
	}

	projectDir := (*tmpCloudManager).GetPredictiveScoringProjectDir(projectId)
	fileDir := (*modelCloudManager).GetPredictiveScoringDir(projectId, startTimestamp, endTimestamp, lookbackWindow)

	if createNewProps {
		err = recreatePropertiesMap(projectId, projectDir, fileDir, dataFilePath, dataFileName, tmpCloudManager, mapProp)
		if err != nil {
			status["err"] = err.Error()
			logCtx.WithError(err).Error("Failed recreating properties file")
			return status, false
		}
	}

	// get properties map
	propFileName := "propMapFiltered.txt"
	var propMapFiltered map[string]map[string]bool
	{
		fileBytes, err := readFileIntoBytesArray(projectDir, propFileName, tmpCloudManager)
		if err != nil {
			status["err"] = err.Error()
			return status, false
		}
		err = json.Unmarshal(fileBytes, &propMapFiltered)
		if err != nil {
			log.WithFields(log.Fields{"fileDir": projectDir, "fileName": propFileName}).Error("Error unmarshalling properties file")
			status["err"] = err.Error()
			return status, false
		}
	}

	// get data to train encoder in python
	encFileName := "encoderTrainData.txt"
	var encoderTrainData map[string][]string
	{
		fileBytes, err := readFileIntoBytesArray(projectDir, encFileName, tmpCloudManager)
		if err != nil {
			status["err"] = err.Error()
			return status, false
		}
		err = json.Unmarshal(fileBytes, &encoderTrainData)
		if err != nil {
			log.WithFields(log.Fields{"fileDir": projectDir, "fileName": encFileName}).Error("Error unmarshalling encoder file")
			status["err"] = err.Error()
			return status, false
		}
	}

	// initialise writer for training data file
	fileName := "training_data.txt"
	cloudWriter, err := (*tmpCloudManager).GetWriter(fileDir, fileName)
	if err != nil {
		log.WithFields(log.Fields{"fileDir": fileDir, "fileName": fileName}).Error("unable to get writer for file")
		status["err"] = err.Error()
		return status, false
	}

	// prepare training data file from data file
	log.Infof("Reading file :%s, %s", dataFilePath, dataFileName)
	file, err := (*tmpCloudManager).Get(dataFilePath, dataFileName)
	if err != nil {
		status["err"] = err.Error()
		return status, false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var dataPoint = make(map[string]interface{})
	var oldId string = "none"
	countLines := 0
	var minTimestamp int64
	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Read failed")
			status["err"] = err.Error()
			return status, false
		}
		AccID := merge.GetAptId(&eventDetails, domain_group.ID)
		if AccID != oldId {
			minTimestamp = eventDetails.EventTimestamp
		}
		dataPoint[idColumnName] = AccID
		eName := eventDetails.EventName
		eventsVals := propMapFiltered["events"]
		if _, ok := eventsVals[eName]; !ok {
			noOfValues := len(eventsVals)
			if noOfValues >= NO_OF_UNIQUE_EVENTS_AND_PROPERTY_VALUES-1 {
				eName = "$others"
			}
			if noOfValues < NO_OF_UNIQUE_EVENTS_AND_PROPERTY_VALUES {

				propMapFiltered["events"][eName] = true
				encoderTrainData[eventsColumnName] = append(encoderTrainData[eventsColumnName], eName)
			}
		}
		dataPoint[eventsColumnName] = eName

		dataPoint["EventCardinality"] = eventDetails.EventCardinality
		dataPoint[timestampColumnName] = eventDetails.EventTimestamp - minTimestamp

		for uKey, uVal := range eventDetails.UserProperties {
			uKey = "up#" + uKey
			if _, ok := propMapFiltered[uKey]; !ok {
				continue
			}
			propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectId, eventDetails.EventName, uKey, uVal, false)
			if propertyType == U.PropertyTypeCategorical {
				strVal := U.GetPropertyValueAsString(uVal)
				if strVal == "" {
					dataPoint[uKey] = strVal
					continue
				}
				valMap := propMapFiltered[uKey]
				if _, ok := valMap[strVal]; !ok {
					noOfValues := len(valMap)
					if noOfValues >= NO_OF_UNIQUE_EVENTS_AND_PROPERTY_VALUES-1 {
						strVal = "$others"
					}
					if noOfValues < NO_OF_UNIQUE_EVENTS_AND_PROPERTY_VALUES {
						propMapFiltered[uKey][strVal] = true
						encoderTrainData[uKey] = append(encoderTrainData[uKey], strVal)
					}
				}
				dataPoint[uKey] = strVal
			} else if propertyType == U.PropertyTypeDateTime {
				intVal, err := U.GetPropertyValueAsInt64(uVal)
				if err != nil {
					log.Error("failed getting interface value")
					status["err"] = err.Error()
					return status, false
				}
				dataPoint[uKey] = intVal - minTimestamp
			} else {
				dataPoint[uKey] = uVal
			}
		}
		for eKey, eVal := range eventDetails.EventProperties {
			eKey = "ep#" + eKey
			if _, ok := propMapFiltered[eKey]; !ok {
				continue
			}
			propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectId, eventDetails.EventName, eKey, eVal, false)
			if propertyType == U.PropertyTypeCategorical {
				strVal := U.GetPropertyValueAsString(eVal)
				if strVal == "" {
					dataPoint[eKey] = strVal
					continue
				}
				valMap := propMapFiltered[eKey]
				if _, ok := valMap[strVal]; !ok {
					noOfValues := len(valMap)
					if noOfValues >= NO_OF_UNIQUE_EVENTS_AND_PROPERTY_VALUES-1 {
						strVal = "$others"
					}
					if noOfValues < NO_OF_UNIQUE_EVENTS_AND_PROPERTY_VALUES {
						propMapFiltered[eKey][strVal] = true
						encoderTrainData[eKey] = append(encoderTrainData[eKey], strVal)
					}
				}
				dataPoint[eKey] = strVal
			} else if propertyType == U.PropertyTypeDateTime {
				intVal, err := U.GetPropertyValueAsInt64(eVal)
				if err != nil {
					log.Error("failed getting interface value")
					status["err"] = err.Error()
					return status, false
				}
				dataPoint[eKey] = intVal - minTimestamp
			} else {
				dataPoint[eKey] = eVal
			}
		}
		if dataPointBytes, err := json.Marshal(dataPoint); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Marshal failed")
			status["err"] = err.Error()
			return status, false
		} else {
			lineWrite := string(dataPointBytes)
			if _, err := io.WriteString(cloudWriter, lineWrite+"\n"); err != nil {
				mineLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
				status["err"] = err.Error()
				return status, false
			}
		}
		oldId = AccID
		countLines++
	}
	err = cloudWriter.Close()
	if err != nil {
		log.WithError(err).Error("error closing writer")
		status["err"] = err.Error()
		return status, false
	}

	// update properties file
	err = CreateFileFromMap(projectDir, propFileName, tmpCloudManager, propMapFiltered)
	if err != nil {
		log.WithError(err).Error("error creating properties file")
		status["err"] = err.Error()
		return status, false
	}
	// update encoder file
	err = CreateFileFromMap(projectDir, encFileName, tmpCloudManager, encoderTrainData)
	if err != nil {
		log.WithError(err).Error("error creating encoder file")
		status["err"] = err.Error()
		return status, false
	}

	return status, true
}

// to delete any previous (properties and encoder) files and create new files using given week
func recreatePropertiesMap(projectId int64, projectDir, fileDir string, cloudPath, cloudName string, cloudManager *filestore.FileManager, filterPropsMap map[string]bool) error {

	var eventsAndPropValsCount = make(map[string]map[string]int)
	log.Infof("Reading file :%s, %s", cloudPath, cloudName)
	file, err := (*cloudManager).Get(cloudPath, cloudName)
	if err != nil {
		log.WithError(err).Error("error reading file")
		return err
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	eventsAndPropValsCount["events"] = make(map[string]int)
	for scanner.Scan() {
		line := scanner.Text()
		var event *P.CounterEventFormat
		err = json.Unmarshal([]byte(line), &event)
		if err != nil {
			log.WithError(err).Error("error unmarshalling line")
			return err
		}
		eventsAndPropValsCount["events"][event.EventName] += 1
		for key, val := range event.EventProperties {
			key = "ep#" + key
			if _, ok := filterPropsMap[key]; !ok {
				continue
			}
			propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectId, event.EventName, key, val, false)
			if propertyType == U.PropertyTypeCategorical {
				strVal := U.GetPropertyValueAsString(val)
				if _, ok := eventsAndPropValsCount[key]; !ok {
					eventsAndPropValsCount[key] = make(map[string]int)
				}
				eventsAndPropValsCount[key][strVal] += 1
			} else if propertyType == U.PropertyTypeDateTime || propertyType == U.PropertyTypeNumerical {
				if _, ok := eventsAndPropValsCount[key]; !ok {
					eventsAndPropValsCount[key] = make(map[string]int)
				}
			}
		}
		for key, val := range event.UserProperties {
			key = "up#" + key
			if _, ok := filterPropsMap[key]; !ok {
				continue
			}
			propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectId, event.EventName, key, val, false)
			if propertyType == U.PropertyTypeCategorical {
				strVal := U.GetPropertyValueAsString(val)
				if _, ok := eventsAndPropValsCount[key]; !ok {
					eventsAndPropValsCount[key] = make(map[string]int)
				}
				eventsAndPropValsCount[key][strVal] += 1
			} else if propertyType == U.PropertyTypeDateTime || propertyType == U.PropertyTypeNumerical {
				if _, ok := eventsAndPropValsCount[key]; !ok {
					eventsAndPropValsCount[key] = make(map[string]int)
				}
			}
		}
	}

	err = CreateFileFromMap(fileDir, "countMap.txt", cloudManager, eventsAndPropValsCount)
	if err != nil {
		log.WithError(err).Error("error creating count file")
		return err
	}

	type Pair struct {
		Key       string
		Frequency int
	}
	var eventsAndPropValsFilt = make(map[string]map[string]bool)
	for key, valMap := range eventsAndPropValsCount {
		// if strings.HasPrefix(key, "ep#$hubspot_") || strings.HasPrefix(key, "up#$hubspot_") {
		// 	continue
		// }
		if len(valMap) == 0 {
			eventsAndPropValsFilt[key] = make(map[string]bool)
			continue
		}

		propsMap := make(map[string]bool)
		var propFreq int
		var uniqueVals int
		freqSlice := make([]Pair, 0)
		if key == "events" {
			for val, freq := range valMap {
				if val == "" {
					continue
				}
				uniqueVals++
				propFreq += freq
			}
			for val, freq := range valMap {
				if val == "" {
					continue
				}
				if freq < propFreq/1000 {
					propsMap["$others"] = true
					continue
				}
				freqSlice = append(freqSlice, Pair{Key: val, Frequency: freq})
			}
		} else {
			for val, freq := range valMap {
				if val == "" {
					continue
				}
				uniqueVals++
				propFreq += freq
				freqSlice = append(freqSlice, Pair{Key: val, Frequency: freq})
			}
			// if float32(uniqueVals)/float32(propFreq) > 0.3 {
			// 	continue
			// }
		}

		sort.Slice(freqSlice, func(i, j int) bool {
			return freqSlice[i].Frequency > freqSlice[j].Frequency
		})
		noOfVals := 0
		for _, pair := range freqSlice {
			noOfVals++
			if noOfVals >= NO_OF_UNIQUE_EVENTS_AND_PROPERTY_VALUES {
				propsMap["$others"] = true
				break
			} else {
				propsMap[pair.Key] = true
			}
		}
		eventsAndPropValsFilt[key] = propsMap
	}
	err = CreateFileFromMap(projectDir, "propMapFiltered.txt", cloudManager, eventsAndPropValsFilt)
	if err != nil {
		log.WithError(err).Error("error creating properties file")
		return err
	}

	var encoderTrainData = make(map[string][]string)
	for key, valMap := range eventsAndPropValsFilt {
		if key == "events" {
			key = eventsColumnName
		}
		if len(valMap) == 0 {
			continue
		}
		if _, ok := encoderTrainData[key]; !ok {
			encoderTrainData[key] = make([]string, 0)
			encoderTrainData[key] = append(encoderTrainData[key], "$others")
		}
		for val, _ := range valMap {
			if val == "" || val == "$others" {
				continue
			}
			encoderTrainData[key] = append(encoderTrainData[key], val)
		}
	}
	err = CreateFileFromMap(projectDir, "encoderTrainData.txt", cloudManager, encoderTrainData)
	if err != nil {
		log.WithError(err).Error("error creating encoding file")
		return err
	}
	return nil
}

// read file into bytes to unmarshal
func readFileIntoBytesArray(fileDir string, fileName string, cloudManager *filestore.FileManager) ([]byte, error) {
	var buffer bytes.Buffer
	file, err := (*cloudManager).Get(fileDir, fileName)
	if err != nil {
		log.WithError(err).Error("error reading file")
		return nil, err
	}
	_, err = buffer.ReadFrom(file)
	if err != nil {
		log.WithFields(log.Fields{"fileDir": fileDir, "fileName": fileName}).Error("Error reading map from File")
		return nil, err
	}
	err = file.Close()
	if err != nil {
		log.WithFields(log.Fields{"fileDir": fileDir, "fileName": fileName}).Error("Error closing file")
		return nil, err
	}
	return buffer.Bytes(), nil
}

// get list of ids from events file
func getIdsFromEventsFile(fileDir, fileName string, cloudManager *filestore.FileManager, idNum int) (map[string]bool, error) {
	var accountIds = make(map[string]bool)
	log.Infof("Reading file :%s, %s", fileDir, fileName)
	file, err := (*cloudManager).Get(fileDir, fileName)
	if err != nil {
		log.WithError(err).Error("error reading file")
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	for scanner.Scan() {
		line := scanner.Text()
		var event *P.CounterEventFormat
		err = json.Unmarshal([]byte(line), &event)
		if err != nil {
			log.WithError(err).Error("error unmarshalling line")
			return nil, err
		}
		accId := merge.GetAptId(event, idNum)
		if accId != "" {
			accountIds[accId] = true
		}
	}
	return accountIds, nil
}
